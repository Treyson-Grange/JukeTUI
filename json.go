package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

// JSON.go
// This file contains the functions for reading and writing JSON data.
// Specifically, it contains the functionality for reading and writing our favorite library items to a JSON file.
// There will be 2 files. One for albums, one for playlists.

func readJSON(filePath string) []LibraryFavorite {
	file, err := os.Open(filePath)
	if err != nil {
		errorLogger.Println("Failed to opsen albums.json: ", err)
		return nil
	}
	defer file.Close()

	favorites := []LibraryFavorite{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&favorites); err != nil {
		log.Fatal("Error while decoding JSON: ", err)
	}

	return favorites
}

func writeJSONFile(filePath string, favorite LibraryFavorite) bool {
	file, err := os.Open(filePath)
	if err != nil {
		errorLogger.Println("Failed to open albums.json: ", err)
		return false
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		errorLogger.Println("Failed to read albums.json: ", err)
		return false
	}

	var favorites []LibraryFavorite
	if err := json.Unmarshal(data, &favorites); err != nil {
		errorLogger.Println("Failed to unmarshal JSON: ", err)
		return false
	}

	favorites = append(favorites, favorite)
	updatedData, err := json.MarshalIndent(favorites, "", "  ")
	if err != nil {
		log.Fatal("failed to marshal JSON: %w", err)
		return false
	}

	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
		log.Fatal("failed to write to file: %w", err)
		return false
	}

	return true
}
