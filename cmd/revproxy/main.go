package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Neffi42/my-http-reverse-proxy/internal/cache"
	"github.com/Neffi42/my-http-reverse-proxy/internal/revproxy"
)

func main() {
	listen := flag.String("listen", ":8000", "address to listen on")
	upstream := flag.String("upstream", "http://127.0.0.1:9000", "address upstream is listening on")
	flag.Parse()

	logger := slog.Default()

	rp, err := revproxy.New(*upstream, nil, logger)
	if err != nil {
		logger.Error("building reverse proxy", "err", err)
		os.Exit(1)
	}

	cacheStore := cache.NewStore()
	cacheMiddleware := cache.New(rp, cacheStore, 5*time.Second, logger)

	logger.Info("proxy listening", "addr", *listen, "upstream", *upstream)

	if err := http.ListenAndServe(*listen, cacheMiddleware); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
