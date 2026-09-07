package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Counter struct {
	mu    sync.Mutex
	count int
}

func hello(port uint) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "hello from port %d\n", port)
	}
}

func testCacheControl(directive string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", directive)
		fmt.Fprintf(w, "testing Cache-Control directive: %s\n", directive)
	}
}

func testHeader(header string, value string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set(header, value)
		fmt.Fprintf(w, "testing Header: %s\n", header)
	}
}

func (c *Counter) counter(w http.ResponseWriter, req *http.Request) {
	c.mu.Lock()
	c.count++
	n := c.count
	c.mu.Unlock()
	fmt.Fprintf(w, "count is: %d\n", n)
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	defaultPort := flag.Uint("port", 9000, "port to listen on")
	replicas := flag.Uint("replicas", 3, "number of backends to run")
	flag.Parse()

	var wg sync.WaitGroup

	for i := range *replicas {
		port := *defaultPort + i
		addr := fmt.Sprintf(":%d", port)
		c := &Counter{}

		mux := http.NewServeMux()
		mux.HandleFunc("/hello", hello(port))
		mux.HandleFunc("/headers", headers)
		mux.HandleFunc("/counter", c.counter)
		mux.HandleFunc("/no-store", testCacheControl("no-store"))
		mux.HandleFunc("/private", testCacheControl("private"))
		mux.HandleFunc("/vary", testHeader("Vary", "Accept-Encoding"))
		mux.HandleFunc("/set-cookie", testHeader("Set-Cookie", "session=12345; Path=/"))

		wg.Go(func() {
			log.Printf("listening on %s", addr)
			if err := http.ListenAndServe(addr, mux); !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("server %s: %v", addr, err)
			}
		})
	}

	wg.Wait()
}
