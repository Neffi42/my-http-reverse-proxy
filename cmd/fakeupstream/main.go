package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

type Counter struct {
	count int
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func (c *Counter) counter(w http.ResponseWriter, req *http.Request) {
	c.count++
	fmt.Fprintf(w, "count is: %d\n", c.count)
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	listen := flag.String("listen", ":9000", "address to listen on")
	flag.Parse()

	c := Counter{}

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/counter", c.counter)

	log.Printf("origin listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
