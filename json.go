package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

// =========================================
// ===== json.go | Read and Write JSON =====
// =========================================

// Open file and return file
func openFile(filePath string) (*os.File, bool) {
	file, err := os.Open(filePath)
	if err != nil {
		errorLogger.Println("Failed to open file: ", err)
		return nil, false
	}
	return file, true
}

// Read the JSON file and return a slice of LibraryFavorite structs.
func readJSON(filePath string) ([]LibraryFavorite, bool) {
	file, _ := openFile(filePath)
	defer file.Close()

	favorites := []LibraryFavorite{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&favorites); err != nil {
		return nil, false
	}

	return favorites, true
}

// Use os.WriteFile to write a new favorite to the JSON file.
func writeJSONFile(filePath string, favorite LibraryFavorite) bool {
	file, _ := openFile(filePath)
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

// Use os.WriteFile to remove a favorite from the JSON file.
func removeFromJSON(filePath string, oldFavorite LibraryFavorite) bool {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	var favorites []LibraryFavorite
	if err := json.Unmarshal(fileData, &favorites); err != nil {
		return false
	}
	var updatedFavorites []LibraryFavorite
	for _, f := range favorites {
		if f.Title != oldFavorite.Title || f.Author != oldFavorite.Author || f.URI != oldFavorite.URI {
			updatedFavorites = append(updatedFavorites, f)
		}
	}
	var updatedData []byte
	if len(updatedFavorites) == 0 {
		updatedData = []byte("[]")
	} else {
		updatedData, err = json.MarshalIndent(updatedFavorites, "", "  ")
		if err != nil {
			return false
		}
	}
	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
		return false
	}

	return true
}

// Use os.WriteFile to create an empty JSON file.
func createEmptyJSONFile(filePath string) bool {
	emptyData := []byte("[]")
	if err := os.WriteFile(filePath, emptyData, 0644); err != nil {
		return false
	}
	return true
}
