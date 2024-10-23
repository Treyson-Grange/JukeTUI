package main

// SpotifyHandlers.go
// This file contains the error handling functions for Spotify API requests.

// Generic fetch/get function for Spotify API requests.
func handleGenericFetch[T any](endpoint, accessToken string, queryParams, bodyArgs map[string]string) T {
	data, err := genericFetch[T](endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		errorLogger.Printf("Failed to fetch data from %s: %v", endpoint, err)
		var empty T
		return empty
	}
	return data
}

// Generic PUT function for Spotify API requests.
func handleGenericPut(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	statusCode, err := genericPut(endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		errorLogger.Printf("Failed to put data to %s: %v", endpoint, err)
		return statusCode, err
	}
	return statusCode, nil
}

// Generic POST function for Spotify API requests.
func handleGenericPost(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	statusCode, err := genericPost(endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		errorLogger.Printf("Failed to post data to %s: %v", endpoint, err)
		return statusCode, err
	}
	return statusCode, nil
}
