package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9999"
	SERVER_TYPE = "tcp"
)

// var blockedSites = make(map[string]bool, 1)

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
	fmt.Print(string(body))
	if err != nil {
		log.Fatal(err.Error())
	}
	copyHeader(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		log.Fatal(err.Error())
	}
	// io.Copy(w, response.Body)

	// // u, err := url.Parse(input_url)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(r.URL.RequestURI())
	// fmt.Println(r.URL.Scheme)
	// fmt.Println(r.URL.Host)
	// fmt.Println(r.Host)
	// // fmt.Println(r.P)
	// fmt.Println(r.URL.Path)
	// processHttpRequest(r.URL)
	// all is the same just has to change the port
	// fmt.Println(u.Scheme)
	// fmt.Println(u.User)
	// fmt.Println(u.Hostname())
	// fmt.Println(u.Port())
	// fmt.Println(u.Path)
	// fmt.Println(u.RawQuery)
	// fmt.Println(u.Fragment)
	// fmt.Println(u.String())
}

func main() {
	http.HandleFunc("/", processURL)
	http.ListenAndServe(":9999", nil)
}
