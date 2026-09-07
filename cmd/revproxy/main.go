package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Neffi42/my-http-reverse-proxy/internal/cache"
	"github.com/Neffi42/my-http-reverse-proxy/internal/revproxy"
)

func main() {
	listen := flag.String("listen", ":8000", "address to listen on")
	flag.Parse()

	routes := map[string]string{
		"/api/":  "http://127.0.0.1:9001",
		"/site/": "http://127.0.0.1:9002",
	}

	logger := slog.Default()
	store := cache.NewStore()

	mux := http.NewServeMux()
	for prefix, upstream := range routes {
		rp, err := revproxy.New(upstream, nil, logger)
		if err != nil {
			logger.Error("building reverse proxy", "prefix", prefix, "err", err)
			os.Exit(1)
		}
		strippedPreffix := http.StripPrefix(strings.TrimSuffix(prefix, "/"), rp)
		cacheMiddleware := cache.New(strippedPreffix, store, 5*time.Second, logger)
		mux.Handle(prefix, cacheMiddleware)
	}

	logger.Info("revproxy listening", "addr", *listen)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
