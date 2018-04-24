package main

import (
	"net/http"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"log"
)


type Response struct {
	Host    string
	UrlPath string
	Scheme	string
}

type sendRequestUrl struct {
}

func (m *sendRequestUrl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var scheme string

	switch {
	case r.URL.Scheme == "https":
		scheme = "https"
	case r.TLS != nil:
		scheme = "https"
	case strings.HasPrefix(r.Proto, "HTTPS"):
		scheme = "https"
	case r.Header.Get("X-Forwarded-Proto") == "https":
		scheme = "https"
	default:
		scheme = "http"
	}
	
	response := Response{r.Host, r.URL.RequestURI(), scheme}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {

	go func() {
		http.ListenAndServe(":80", &sendRequestUrl{})
	}()

	dir, cerr := os.Getwd()
	if cerr != nil {
		log.Fatal(cerr)
	}
	fmt.Println(dir)

	err := http.ListenAndServeTLS(":443", "server.cert", "server.key", &sendRequestUrl{})
	log.Fatal(err)
}
