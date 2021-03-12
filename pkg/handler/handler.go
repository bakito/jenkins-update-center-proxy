package handler

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bakito/jenkins-update-center-proxy/version"
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

func New(contextPath string, repoProxyURL string, offlineDir string) *mux.Router {
	cp := contextPath
	if cp == "/" {
		cp = ""
	}

	h := &handler{
		repoProxyURL: repoProxyURL,
		contextPath:  cp,
		offlineFiles: make(map[string][]byte),
	}

	h.readOfflineFiles(offlineDir)

	r := mux.NewRouter()
	r.HandleFunc(contextPath, h.handleIndex)
	r.HandleFunc(cp+updateCenter, h.handleUpdateCenter(updateCenter))
	r.HandleFunc(cp+updateCenterActual, h.handleUpdateCenter(updateCenterActual))
	r.HandleFunc(cp+updateCenterExperimental, h.handleUpdateCenter(updateCenterExperimental))
	r.HandleFunc(cp+pluginVersions, h.handleUpdateCenter(pluginVersions))
	return r
}

type handler struct {
	repoProxyURL string
	offlineFiles map[string][]byte
	contextPath  string
}

func (h *handler) readOfflineFiles(offlineDir string) {
	if offlineDir != "" {
		h.loadFile(offlineDir, updateCenter)
		h.loadFile(offlineDir, updateCenterActual)
		h.loadFile(offlineDir, updateCenterExperimental)
		h.loadFile(offlineDir, pluginVersions)
	}
}

func (h *handler) loadFile(offlineDir string, name string) {
	path := filepath.Join(offlineDir, name)
	if _, err := os.Stat(path); err == nil {
		if dat, err := ioutil.ReadFile(path); err == nil {

			ucj := string(dat)
			ucj = strings.ReplaceAll(ucj, repoURL, h.repoProxyURL)
			h.offlineFiles[name] = []byte(ucj)
			fmt.Printf("Using offline file %s\n", path)
		}
	}
}

// Given a request send it to the appropriate url
func (h *handler) handleUpdateCenter(file string) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {

		url := repoURL + file

		fmt.Printf("Request %s\n", url)

		var dat []byte
		if content, ok := h.offlineFiles[file]; ok {
			dat = content
		} else {
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
