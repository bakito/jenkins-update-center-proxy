package handler

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
)

const (
	repoURL                     = "https://updates.jenkins.io"
	updateCenter                = "/update-center.json"
	updateCenterActual          = "/update-center.actual.json"
	updateCenterExperimental    = "/experimental/update-center.json"
	updateCenterURL             = repoURL + updateCenter
	updateCenterActualURL       = repoURL + updateCenterActual
	updateCenterExperimentalURL = repoURL + updateCenterExperimental
)

var (
	//go:embed index.html
	index []byte
)

func New(repoProxyURL string) *mux.Router {
	h := &handler{
		repoProxyURL: repoProxyURL,
	}
	r := mux.NewRouter()
	r.HandleFunc("/", h.handleIndex)
	r.HandleFunc(updateCenter, h.handleUpdateCenter(updateCenterURL))
	r.HandleFunc(updateCenterActual, h.handleUpdateCenter(updateCenterActualURL))
	r.HandleFunc(updateCenterExperimental, h.handleUpdateCenter(updateCenterExperimentalURL))
	return r
}

type handler struct {
	repoProxyURL string
}

// Given a request send it to the appropriate url
func (h *handler) handleUpdateCenter(url string) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("Request %s\n", url)
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
		res.Header().Set("Content-Type", resp.Header().Get("Content-Type"))
		res.Header().Set("Content-Length", fmt.Sprintf("%d", len([]byte(ucj))))
		_, err = res.Write([]byte(ucj))
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
