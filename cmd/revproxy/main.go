package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Neffi42/my-http-reverse-proxy/internal/revproxy"
)

func main() {
	listen := flag.String("listen", ":8000", "address to listen on")
	origin := flag.String("origin", "http://127.0.0.1:9000", "address origin is listening on")
	flag.Parse()

	rp, err := revproxy.New(*origin)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("proxy on %s -> %s\n", *listen, *origin)
	log.Fatal(http.ListenAndServe(*listen, rp))
}
