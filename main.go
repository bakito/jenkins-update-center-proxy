package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bakito/jenkins-update-center-proxy/pkg/handler"
	"github.com/bakito/jenkins-update-center-proxy/version"
)

const (
	envRepoProxyURL = "REPO_PROXY_URL"
	envPort         = "PORT"
	envOfflineDir   = "OFFLINE_DIR"
	envContextPath  = "CONTEXT_PATH"
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

	fmt.Printf("Starting server %s on port %s\n", version.Version, port)
	contextPath := "/"
	if cp, ok := os.LookupEnv(envContextPath); ok {
		if !strings.HasPrefix(cp, "/") {
			cp = "/" + cp
		}
		if strings.HasSuffix(cp, "/") {
			cp = cp[:len(cp)-1]
		}
		contextPath = cp
		fmt.Printf("Context path is: %s\n", contextPath)
	}

	offlineDir := os.Getenv(envOfflineDir)

	router := handler.New(contextPath, repoProxyURL, offlineDir)

	http.Handle("/", router)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		panic(err)
	}
}
