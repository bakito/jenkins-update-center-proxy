package handler

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
)

const (
	repoURL = "https://updates.jenkins.io"

	updateCenter       = "/update-center.json"
	updateCenterActual = "/update-center.actual.json"
)

var (
	//go:embed index.html
	index []byte
)

func New(repoProxyURL string, offlineDir string) *mux.Router {
	h := &handler{
		repoProxyURL: repoProxyURL,
		offlineFiles: make(map[string][]byte),
	}

	h.readOfflineFiles(offlineDir)

	r := mux.NewRouter()
	r.HandleFunc("/", h.handleIndex)
	r.HandleFunc(updateCenter, h.handleUpdateCenter(updateCenter))
	r.HandleFunc(updateCenterActual, h.handleUpdateCenter(updateCenterActual))
	return r
}

type handler struct {
	repoProxyURL string
	offlineFiles map[string][]byte
}

func (h *handler) readOfflineFiles(offlineDir string) {
	if offlineDir != "" {
		h.loadFile(offlineDir, updateCenter)
		h.loadFile(offlineDir, updateCenterActual)
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

func (h *handler) handleIndex(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html")
	_, err := res.Write(index)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
