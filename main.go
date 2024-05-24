package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

// Target server URL
var targetURL = "http://example.com"

// handleRequestAndRedirect handles the incoming request and redirects it to the target server
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {

	// Parse the target URL
	url, err := url.Parse(targetURL)
	if err != nil {
		http.Error(res, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	// Create a new request based on the incoming request
	proxyReq, err := http.NewRequest(req.Method, url.String(), req.Body)
	if err != nil {
		http.Error(res, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	// Copy the headers from the incoming request to the proxy request
	proxyReq.Header = req.Header

	// Make the request to the target server
	client := &http.Client{}
	proxyRes, err := client.Do(proxyReq)
	if err != nil {
		http.Error(res, "Error making request to target server", http.StatusInternalServerError)
		return
	}
	defer proxyRes.Body.Close()

	// Copy the headers from the target server response to the proxy response
	for key, value := range proxyRes.Header {
		res.Header()[key] = value
	}

	// Write the status code from the target server response
	res.WriteHeader(proxyRes.StatusCode)

	// log the request
	log.Printf("Request: %s %s %s\n", req.Method, req.URL, req.Proto)

	// Copy the body from the target server response to the proxy response
	io.Copy(res, proxyRes.Body)
}

func main() {
	http.HandleFunc("/", handleRequestAndRedirect)
	log.Println("Starting proxy server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting proxy server: %v", err)
	}
}
