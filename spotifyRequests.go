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
// Should rename this file eventually.

// genericFetch makes a GET request to the Spotify API and returns the response as a struct.
//
// Parameters:
// - endpoint: the endpoint to fetch data from
// - accessToken: the access token to authenticate the request
// - queryParams: the query parameters to include in the request
// - bodyArgs: the body arguments to include in the request
//
// Returns:
// - T: the response data as a struct
// - error: an error if the request fails
//
// Type Parameters:
// - T: the type of the response data
func genericFetch[T any](endpoint, accessToken string, queryParams, bodyArgs map[string]string) (T, error) {
	var result T

	req, err := http.NewRequest(http.MethodGet, createEndpoint(endpoint, queryParams), nil)
	if err != nil {
		return result, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	if bodyArgs != nil {
		body, err := json.Marshal(bodyArgs)
		if err != nil {
			return result, err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return result, fmt.Errorf("missing required permissions")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}

// genericPut makes a PUT request to the Spotify API and returns the response code.
//
// Parameters:
// - endpoint: the endpoint to fetch data from
// - accessToken: the access token to authenticate the request
// - queryParams: the query parameters to include in the request
// - bodyArgs: the body arguments to include in the request
//
// Returns:
// - int: the response code
// - error: an error if the request fails
func genericPut(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	body, err := json.Marshal(bodyArgs)
	if err != nil {
		return 500, err
	}

	req, err := http.NewRequest(http.MethodPut, createEndpoint(endpoint, queryParams), io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return 500, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			return resp.StatusCode, err
		}
		return 500, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// genericPost makes a POST request to the Spotify API and returns the response code.
//
// Parameters:
// - endpoint: the endpoint to fetch data from
// - accessToken: the access token to authenticate the request
// - queryParams: the query parameters to include in the request
// - bodyArgs: the body arguments to include in the request
//
// Returns:
// - int: the response code
// - error: an error if the request fails
func genericPost(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	body, err := json.Marshal(bodyArgs)
	if err != nil {
		return 500, err
	}

	req, err := http.NewRequest(http.MethodPost, createEndpoint(endpoint, queryParams), io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return 500, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			return resp.StatusCode, err
		}
		return 500, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}
