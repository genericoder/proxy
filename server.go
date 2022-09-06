package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9999"
	SERVER_TYPE = "tcp"
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
		log.Fatal(err.Error())
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	copyHeader(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	http.HandleFunc("/", processURL)
	http.ListenAndServe(":9999", nil)
}
