package main

import (
	"context"
	"fmt"
	"github.com/rueian/rueidis"
	"github.com/rueian/rueidis/rueidiscompat"
	"log"
	"net/http"
	"time"
)

// defaultTtl TTL is 1 hour. Needed in case we lose our connection and then miss invalidation events
const defaultTtl = time.Hour

type redirector struct {
	cache rCache
}

type rCache interface {
	lookup(ctx context.Context, key string) (string, error)
}

type redisSimple struct {
	cclient rueidiscompat.Cmdable
}

func (r redisSimple) lookup(ctx context.Context, key string) (string, error) {
	// looks up through the local client cache:
	res, err := r.cclient.Cache(defaultTtl).Get(ctx, key).Result()
	// cache misses are errors, handle them gracefully.
	if rueidis.IsRedisNil(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("lookup error: %w", err)
	}
	return res, nil
}

// HTTP Handler for redirects
func (r redirector) redirectHandler(w http.ResponseWriter, req *http.Request) {
	ctx := context.TODO()
	key, err := getFullUrl(req)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	val, err := r.cache.lookup(ctx, key)
	if err != nil {
		log.Println("lookup: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if val == "" {
		log.Println("Url not found", key)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	log.Println("Redirecting", key, "to", val)
	http.Redirect(w, req, val, http.StatusFound)

}

// Takes a request and returns the complete URL
func getFullUrl(req *http.Request) (string, error) {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + req.Host + req.URL.Path, nil
}

func main() {
	client, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{"127.0.0.1:6379"}})
	if err != nil {
		panic(err)
	}
	defer client.Close()
	compat := rueidiscompat.NewAdapter(client)
	populate(compat)

	srv := &http.Server{
		Addr: ":8080",
	}
	urlCache := redisSimple{cclient: compat}
	r := redirector{cache: urlCache}
	http.HandleFunc("/", r.redirectHandler)
	log.Fatal(srv.ListenAndServe())
}

func populate(compat rueidiscompat.Cmdable) {
	ctx := context.TODO()
	res := compat.SetNX(ctx, "http://localhost:8080/", "https://www.google.com/", defaultTtl)
	if res.Err() != nil {
		panic(res.Err())
	}
	res = compat.SetNX(ctx, "http://localhost:8080/yahoo", "https://www.yahoo.com/", defaultTtl)
	if res.Err() != nil {
		panic(res.Err())
	}

}
