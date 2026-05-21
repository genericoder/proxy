package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func processURL(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get(r.URL.String())
	if err != nil {
		log.Printf("upstream request failed: %s", err.Error())
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("reading upstream body failed: %s", err.Error())
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}
	copyHeader(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		log.Printf("writing response failed: %s", err.Error())
	}
}

func listenAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":19998"
}

func main() {
	http.HandleFunc("/", processURL)
	http.ListenAndServe(listenAddr(), nil)
}
