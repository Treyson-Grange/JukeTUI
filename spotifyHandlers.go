package main

import (
	"log"
)

// SpotifyHandlers.go
// This file contains the error handling functions for Spotify API requests.

// Generic fetch/get function for Spotify API requests.
func handleGenericFetch[T any](endpoint, accessToken string, queryParams map[string]string) T {
	data, err := genericFetch[T](endpoint, accessToken, queryParams)
	if err != nil {
		log.Fatalf("Failed to fetch data: %v", err)
	}
	return data
}

// Generic PUT function for Spotify API requests.
func handleGenericPut(endpoint, accessToken string, queryParams map[string]string) (int, error) {
	statusCode, err := genericPut(endpoint, accessToken, queryParams)
	if err != nil {
		log.Fatalf("Failed to send PUT request: %v", err)
	}
	return statusCode, nil
}

// Generic POST function for Spotify API requests.
func handleGenericPost(endpoint, accessToken string, queryParams map[string]string) (int, error) {
	statusCode, err := genericPost(endpoint, accessToken, queryParams)
	if err != nil {
		log.Fatalf("Failed to send PUT request: %v", err)
	}
	return statusCode, nil
}
