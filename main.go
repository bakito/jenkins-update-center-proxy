package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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
	envTimeout                 = "TIMEOUT"
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

	timeoutString := "1m"
	if to, ok := os.LookupEnv(envTimeout); ok {
		timeoutString = to
	}

	timeout, err := time.ParseDuration(timeoutString)
	if err != nil {
		log.Error("timeout %q from env var %q is invalid", timeoutString, envTimeout)
		os.Exit(1)
	}

	log.With("version", version.Version, "port", port, "contextPath", contextPath).Info("Starting server")
	useProxyForDownload := strings.EqualFold("true", os.Getenv(envUseRepoProxyForDownload))
	insecureSkipVerify := strings.EqualFold("true", os.Getenv(envInsecureSkipVerify))

	offlineDir := os.Getenv(envOfflineDir)
	r := mux.NewRouter()
	h := handler.New(r, contextPath, repoProxyURL, useProxyForDownload, insecureSkipVerify, offlineDir, timeout)
	defer h.Close()

	http.Handle("/", r)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		panic(err)
	}
}
