package handler

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bakito/jenkins-update-center-proxy/version"
	"github.com/fsnotify/fsnotify"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
)

const (
	repoURL = "https://updates.jenkins.io"

	updateCenter             = "/update-center.json"
	updateCenterActual       = "/update-center.actual.json"
	updateCenterExperimental = "/experimental/update-center.json"
	pluginVersions           = "/plugin-versions.json"
)

var (
	//go:embed index.html
	indexTemplate string
)

func New(r *mux.Router, contextPath string, repoProxyURL string, useProxyForDownload bool, offlineDir string) Handler {
	cp := contextPath
	if cp == "/" {
		cp = ""
	}

	h := &handler{
		repoProxyURL: repoProxyURL,
		offlineDir:   offlineDir,
		contextPath:  cp,
		offlineFiles: make(map[string]string),
	}

	if useProxyForDownload {
		h.downloadURL = repoProxyURL
	} else {
		h.downloadURL = repoURL
	}

	if h.offlineDir != "" {
		var err error
		h.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}

		if err := h.watcher.Add(h.offlineDir); err != nil {
			log.Printf("ERROR: %v", err)
		}
		go h.watchOfflineChanges()
	}

	r.HandleFunc(contextPath, h.handleIndex)
	r.HandleFunc(cp+updateCenter, h.handleUpdateCenter(updateCenter))
	r.HandleFunc(cp+updateCenterActual, h.handleUpdateCenter(updateCenterActual))
	r.HandleFunc(cp+updateCenterExperimental, h.handleUpdateCenter(updateCenterExperimental))
	r.HandleFunc(cp+pluginVersions, h.handleUpdateCenter(pluginVersions))

	h.readOfflineFiles()
	return h
}

type Handler interface {
	Close()
}

type handler struct {
	repoProxyURL string
	downloadURL  string
	offlineFiles map[string]string
	contextPath  string
	offlineDir   string
	watcher      *fsnotify.Watcher
}

func (h *handler) Close() {
	if h.watcher != nil {
		_ = h.watcher.Close()
	}
}

func (h *handler) readOfflineFiles() {
	if h.offlineDir != "" {
		log.Printf("Reload offline files %s\n", h.offlineDir)
		h.cacheFile(h.offlineDir, updateCenter)
		h.cacheFile(h.offlineDir, updateCenterActual)
		h.cacheFile(h.offlineDir, updateCenterExperimental)
		h.cacheFile(h.offlineDir, pluginVersions)
	}
}

func (h *handler) watchOfflineChanges() {
	for {
		select {
		// watch for events
		case event := <-h.watcher.Events:
			if strings.HasSuffix(event.Name, ".json") {
				h.readOfflineFiles()
			}

			// watch for errors
		case err := <-h.watcher.Errors:
			log.Printf("ERROR: %v", err)
		}
	}
}
func (h *handler) cacheFile(offlineDir string, name string) {
	path := filepath.Join(offlineDir, name)
	dat := h.loadFile(path)
	if dat != nil {
		h.offlineFiles[name] = path
	} else {
		delete(h.offlineFiles, name)
	}
}

func (h *handler) loadFile(path string) []byte {
	if _, err := os.Stat(path); err == nil {
		if dat, err := ioutil.ReadFile(path); err == nil {
			ucj := string(dat)
			ucj = strings.ReplaceAll(ucj, repoURL, h.repoProxyURL)
			return []byte(ucj)
		}
	}
	return nil
}

// Given a request send it to the appropriate url
func (h *handler) handleUpdateCenter(file string) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {

		url := h.downloadURL + file

		if req.URL.RawQuery != "" {
			log.Printf("Request %s?%s\n", url, req.URL.RawQuery)
		} else {
			log.Printf("Request %s\n", url)
		}

		var dat []byte
		if path, ok := h.offlineFiles[file]; ok {
			dat = h.loadFile(path)
		}

		if dat == nil {
			rc := resty.New()
			rc.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

			resp, err := rc.R().
				SetQueryParamsFromValues(req.URL.Query()).
				EnableTrace().
				Get(url)

			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			ucj := string(resp.Body())

			ucj = strings.ReplaceAll(ucj, repoURL, h.repoProxyURL)
			dat = []byte(ucj)
		}

		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("Content-Length", fmt.Sprintf("%d", len(dat)))
		_, err := res.Write(dat)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) handleIndex(res http.ResponseWriter, _ *http.Request) {

	t, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"title":       fmt.Sprintf("Jenkins UpdateCenter Proxy %s", version.Version),
		"contextPath": h.contextPath,
		"links": []string{
			updateCenter,
			updateCenterActual,
			updateCenterExperimental,
			pluginVersions,
		},
	}

	res.Header().Set("Content-Type", "text/html")
	err = t.Execute(res, data)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
