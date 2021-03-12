package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

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
		log.Printf("env variable %s is required", envRepoProxyURL)
		os.Exit(1)
	}
	port := "8080"
	if p, ok := os.LookupEnv(envPort); ok {
		port = p
	}

	log.Printf("Starting server %s on port %s\n", version.Version, port)
	contextPath := "/"
	if cp, ok := os.LookupEnv(envContextPath); ok {
		if !strings.HasPrefix(cp, "/") {
			cp = "/" + cp
		}
		if strings.HasSuffix(cp, "/") {
			cp = cp[:len(cp)-1]
		}
		contextPath = cp
		log.Printf("Context path is: %s\n", contextPath)
	}

	offlineDir := os.Getenv(envOfflineDir)
	r := mux.NewRouter()
	h := handler.New(r, contextPath, repoProxyURL, offlineDir)
	defer h.Close()

	http.Handle("/", r)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		panic(err)
	}
}
