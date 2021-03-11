package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/allegro/bigcache"

	"github.com/bakito/jenkins-update-center-proxy/pkg/handler"
	"github.com/eko/gocache/store"
)

const (
	envRepoProxyURL = "REPO_PROXY_URL"
	envPort         = "PORT"
)

func main() {

	repoProxyURL := os.Getenv(envRepoProxyURL)
	if repoProxyURL == "" {
		fmt.Printf("env variable %s is required", envRepoProxyURL)
		os.Exit(1)
	}
	port := "8080"
	if p, ok := os.LookupEnv(envPort); ok {
		port = p
	}

	bigcacheClient, _ := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	bigcacheStore := store.NewBigcache(bigcacheClient, nil)

	router := handler.New(bigcacheStore, repoProxyURL)

	http.Handle("/", router)
	fmt.Printf("Starting servier on port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		panic(err)
	}
}
