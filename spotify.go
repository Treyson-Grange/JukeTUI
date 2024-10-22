package main

import (
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

func genericFetch[T any](endpoint, accessToken string, queryParams map[string]string) (T, error) {
	var result T

	req, err := http.NewRequest(http.MethodGet, createEndpoint(endpoint, queryParams), nil)
	if err != nil {
		return result, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

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

func genericPut(endpoint, accessToken string, queryParams map[string]string) (int, error) {
	req, err := http.NewRequest(http.MethodPut, createEndpoint(endpoint, queryParams), nil)
	if err != nil {
		return 500, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return resp.StatusCode, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func genericPost(endpoint, accessToken string, queryParams map[string]string) (int, error) {
	req, err := http.NewRequest(http.MethodPost, createEndpoint(endpoint, queryParams), nil)
	if err != nil {
		return 500, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return resp.StatusCode, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}
