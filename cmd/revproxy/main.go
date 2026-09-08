package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Neffi42/my-http-reverse-proxy/internal/cache"
	"github.com/Neffi42/my-http-reverse-proxy/internal/config"
	"github.com/Neffi42/my-http-reverse-proxy/internal/revproxy"
)

func main() {
	configPath := flag.String("config", "", "path to config file (required)")
	flag.Parse()

	if *configPath == "" {
		flag.Usage()
		os.Exit(2)
	}

	logger := slog.Default()

	logger.Info("loading config", "config", *configPath)
	config, err := config.New(*configPath)
	if err != nil {
		logger.Error("loading config", "err", err)
		os.Exit(1)
	}

	store := cache.NewStore()
	mux := http.NewServeMux()

	for _, r := range config.Routes {
		logger.Info("building route", "prefix", r.Prefix, "upstream", r.Upstream)
		rp, err := revproxy.New(r.Upstream, nil, logger)
		if err != nil {
			logger.Error("building reverse proxy", "prefix", r.Prefix, "err", err)
			os.Exit(1)
		}
		ttl, err := time.ParseDuration(r.Ttl)
		if err != nil {
			logger.Error("parsing route ttl", "prefix", r.Prefix, "ttl", r.Ttl, "err", err)
			os.Exit(1)
		}
		strippedPreffix := http.StripPrefix(strings.TrimSuffix(r.Prefix, "/"), rp)
		cacheMiddleware := cache.New(strippedPreffix, store, ttl, logger)
		mux.Handle(r.Prefix, cacheMiddleware)
	}

	logger.Info("revproxy ready and listening", "addr", config.Listen)
	if err := http.ListenAndServe(config.Listen, mux); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
