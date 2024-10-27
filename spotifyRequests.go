package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Spotify.go
// This file contains the generic functions for fetching data from the Spotify API.
// These functions are used by the handlers to fetch data from the Spotify API.

// genericRequest makes an HTTP request to the Spotify API and returns the response as a struct or a response code.
//
// Parameters:
// - method: the HTTP method to use (GET, POST, PUT)
// - endpoint: the endpoint to fetch data from
// - accessToken: the access token to authenticate the request
// - queryParams: the query parameters to include in the request
// - bodyArgs: the body arguments to include in the request
//
// Returns:
// - T: the response data as a struct if method is GET
// - int: the response code if method is POST or PUT
// - error: an error if the request fails
//
// Type Parameters:
// - T: the type of the response data
func genericRequest[T any](method, endpoint, accessToken string, queryParams, bodyArgs map[string]string) (T, int, error) {
	var result T
	var resp *http.Response

	// Create request body if bodyArgs are provided
	var body io.Reader
	if bodyArgs != nil {
		bodyJSON, err := json.Marshal(bodyArgs)
		if err != nil {
			return result, 500, err
		}
		body = bytes.NewReader(bodyJSON)
	}

	req, err := http.NewRequest(method, createEndpoint(endpoint, queryParams), body)
	if err != nil {
		return result, 500, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		if resp != nil {
			return result, resp.StatusCode, err
		}
		return result, 500, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return result, resp.StatusCode, fmt.Errorf("missing required permissions")
	}

	if method == http.MethodGet {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return result, resp.StatusCode, err
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return result, resp.StatusCode, err
		}
	}

	return result, resp.StatusCode, nil
}

// genericFetch makes a GET request to the Spotify API and returns the response as a struct.
func genericFetch[T any](endpoint, accessToken string, queryParams, bodyArgs map[string]string) (T, error) {
	result, _, err := genericRequest[T](http.MethodGet, endpoint, accessToken, queryParams, bodyArgs)
	return result, err
}

// genericPut makes a PUT request to the Spotify API and returns the response code.
func genericPut(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	_, statusCode, err := genericRequest[struct{}](http.MethodPut, endpoint, accessToken, queryParams, bodyArgs)
	return statusCode, err
}

// genericPost makes a POST request to the Spotify API and returns the response code.
func genericPost(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	_, statusCode, err := genericRequest[struct{}](http.MethodPost, endpoint, accessToken, queryParams, bodyArgs)
	return statusCode, err
}
