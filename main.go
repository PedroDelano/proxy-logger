package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Target server URL
var targetURL = "http://localhost:8000"

// handleRequestAndRedirect handles the incoming request and redirects it to the target server
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {

	// this will join the specific path with the targetURL
	// Example:
	// - targetURL = http://localhost:8000
	// - req.URL.Path = /api/v1/users
	// - customURL = http://localhost:8000/api/v1/users
	customURL := targetURL + req.URL.Path

	// Parse the target URL
	url, err := url.Parse(customURL)
	if err != nil {
		http.Error(res, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	// Read and save the request body
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	// Create a new proxy request with the parsed URL and request body
	proxyReq, err := http.NewRequest(req.Method, url.String(), bytes.NewReader(reqBody))
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

	// Read and save the response body
	resBody, err := io.ReadAll(proxyRes.Body)
	if err != nil {
		http.Error(res, "Error reading response body", http.StatusInternalServerError)
		return
	}

	// log the request
	log.Printf("Request: %s %s %s\n", req.Method, req.URL, req.Proto)

	// Transform reqBody and resBody into json
	jsonReqBody := string(reqBody)
	jsonResBody := string(resBody)

	log.Println("jsonReqBody", jsonReqBody)
	log.Println("jsonResBody", jsonResBody)

	// Write the response body to the client
	res.Write(resBody)

}

func main() {
	http.HandleFunc("/", handleRequestAndRedirect)
	log.Println("Starting proxy server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting proxy server: %v", err)
	}
}
