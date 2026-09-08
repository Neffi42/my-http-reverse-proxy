# my-http-reverse-proxy

A small hobby HTTP reverse proxy I built.

## Context

It is meant to be a fun practice exercise, not something built for production.
It's inspired by the implementation of `net/http/httputil.ReverseProxy`.
I took reference from:

- [RFC 9110](https://www.rfc-editor.org/info/rfc9110/) for the reverse proxy.
- [RFC 9111](https://www.rfc-editor.org/info/rfc9111/) for the HTTP caching.

## Features

- Routes requests to different upstream backends by URL path prefix (e.g. `/api/` -> one backend, `/private/` -> another, `/` -> default).
- Forwards `X-Forwarded-For`, `X-Forwarded-Proto`, `X-Forwarded-Host` and strips hop-by-hop headers in and out.
- Caches GET responses in memory for a per-route TTL, adding an `X-Cache` header (`MISS` / `HIT` / `BYPASS`) and an `Age` header on hits.
- Safely bypasses the cache for sensitive or negotiated data (respects `Cache-Control: private/no-store`, client `Cookie`, and server `Set-Cookie`/`Vary` headers).

## Project Structure

```text
cmd/
    revproxy/       main binary: loads config, wires routing + cache, listens
    fakeupstream/   throw-away backend for local testing (spins up N ports, each with /hello, /headers, /counter)
internal/
    revproxy/       single-upstream reverse proxy handler (RevProxy)
    cache/          in-memory cache middleware (Middleware, Store, Entry)
    config/         JSON config loading (Config)
```

## Getting Started

I built this with Go _1.27.1_.
I use [mise](https://mise.jdx.dev/) to manage Go's version.
Check it out if you are intersted, it's pretty cool.

`fakeupstream` will probably break if you use a version older than _1.25_ since I use the `sync.WaitGroup.Go()` function.

A JSON config file is required to be passed via the `-config` flag.

```json
{
  "listen": ":8000",
  "routes": [
    {
      "prefix": "/",
      "upstream": "http://localhost:9000",
      "ttl": "5s"
    },
    {
      "prefix": "/api/",
      "upstream": "http://localhost:9001",
      "ttl": "10s"
    },
    {
      "prefix": "/private/",
      "upstream": "http://localhost:9002",
      "ttl": "1m"
    }
  ]
}
```

`listen` is an address for `http.ListenAndServe`.
`routes` is a list of prefix/upstream/ttl entries. `prefix` needs a trailing slash to subtree-match (see `net/http.ServeMux` rules), `upstream` is the backend base URL, and `ttl` (parsed with `time.ParseDuration`) is that route's cache TTL.

You can use [config.example.json](./config.example.json) to test it out with `fakeupstream`.
The config matches `fakeupstream` default options.

If you want you can override the port (`-port`) or the number of backends (`-replicas`) via `fakeupstream` flags, but be careful to update the config to match if you do.

## Demo

First, you'll want to start three fake backends:

```sh
go run ./cmd/fakeupstream
```

Then, in another terminal, start the proxy:

```sh
go run ./cmd/revproxy -config=./config.example.json
```

If you make a request, the first one will be a cache miss and gets forwarded to the upstream backend:

```sh
curl -i http://localhost:8000/api/counter
# HTTP/1.1 200 OK
# X-Cache: MISS
# ...
# count is: 1
```

If you repeat that request within the TTL (10s for `/api/` in the example config), it will be served directly from the cache. You won't hit the upstream, and you can see the `Age` header gets added:

```sh
curl -i http://localhost:8000/api/counter
# HTTP/1.1 200 OK
# X-Cache: HIT
# Age: 1
# ...
# count is: 1
```

Try hitting a different route prefix to reach a different upstream counter. This proves that the routes don't collide in the shared cache:

```sh
curl -i http://localhost:8000/counter
# X-Cache: MISS
# count is: 1
```

Finally, if you wait past the 10-second TTL and make the original request again, you'll see an X-Cache: MISS and the count will increment.

## Known limitations

Since this is just a practice project, I left a few things out:

- **It ignores `Cache-Control` TTL directives (`max-age`/`s-maxage`), relying entirely on the fixed TTL.**
- The Store never actually evicts expired entries, so it will have unbounded memory growth over a long uptime.
- I used http.ListenAndServe directly, which means there are no timeouts or graceful shutdowns configured.
- It only caches GET requests, and there is no cap on the cache size.
- I left a TODO in the proxy code to properly preserve escaped slashes (`%2F`) by joining `EscapedPath()` too.
