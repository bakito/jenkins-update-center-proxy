package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bakito/jenkins-update-center-proxy/pkg/handler"
	"github.com/bakito/jenkins-update-center-proxy/version"
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

	router := handler.New(repoProxyURL)

	http.Handle("/", router)
	fmt.Printf("Starting server %s on port %s\n", version.Version, port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		panic(err)
	}
}
