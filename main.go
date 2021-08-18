package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bakito/jenkins-update-center-proxy/pkg/handler"
	"github.com/bakito/jenkins-update-center-proxy/version"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	envRepoProxyURL            = "REPO_PROXY_URL"
	envUseRepoProxyForDownload = "USE_REPO_PROXY_FOR_DOWNLOAD"
	envPort                    = "PORT"
	envOfflineDir              = "OFFLINE_DIR"
	envContextPath             = "CONTEXT_PATH"
	envInsecureSkipVerify      = "TLS_INSECURE_SKIP_VERIFY"
)

func main() {
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	repoProxyURL := os.Getenv(envRepoProxyURL)
	if repoProxyURL == "" {
		log.Error("env variable %s is required", envRepoProxyURL)
		os.Exit(1)
	}
	port := "8080"
	if p, ok := os.LookupEnv(envPort); ok {
		port = p
	}

	contextPath := "/"
	if cp, ok := os.LookupEnv(envContextPath); ok {
		if !strings.HasPrefix(cp, "/") {
			cp = "/" + cp
		}
		contextPath = strings.TrimSuffix(cp, "/")
	}

	insecureSkipVerify := strings.ToLower(os.Getenv(envInsecureSkipVerify)) == "true"

	log.With("version", version.Version, "port", port, "contextPath", contextPath).Info("Starting server")
	useProxyForDownload := strings.EqualFold("true", os.Getenv(envUseRepoProxyForDownload))

	offlineDir := os.Getenv(envOfflineDir)
	r := mux.NewRouter()
	h := handler.New(r, contextPath, repoProxyURL, useProxyForDownload, insecureSkipVerify, offlineDir)
	defer h.Close()

	http.Handle("/", r)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		panic(err)
	}
}
