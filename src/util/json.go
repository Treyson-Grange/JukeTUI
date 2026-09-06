package util

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/treyson-grange/JukeTUI/src/types"
)

// =========================================
// ===== json.go | Read and Write JSON =====
// =========================================

// Open file and return file
func OpenFile(filePath string) (*os.File, bool) {
	file, err := os.Open(filePath)
	if err != nil {
		ErrorLogger.Println("Failed to open file: ", err)
		return nil, false
	}
	return file, true
}

// Read the JSON file and return a slice of types.LibraryFavorite structs.
func ReadJSON(filePath string) ([]types.LibraryFavorite, bool) {
	file, _ := OpenFile(filePath)
	defer file.Close()

	favorites := []types.LibraryFavorite{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&favorites); err != nil {
		return nil, false
	}

	return favorites, true
}

// Use os.WriteFile to write a new favorite to the JSON file.
func WriteJSONFile(filePath string, favorite types.LibraryFavorite) bool {
	file, _ := OpenFile(filePath)
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		ErrorLogger.Println("Failed to read albums.json: ", err)
		return false
	}

	var favorites []types.LibraryFavorite
	if err := json.Unmarshal(data, &favorites); err != nil {
		ErrorLogger.Println("Failed to unmarshal JSON: ", err)
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
func RemoveFromJSON(filePath string, oldFavorite types.LibraryFavorite) bool {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	var favorites []types.LibraryFavorite
	if err := json.Unmarshal(fileData, &favorites); err != nil {
		return false
	}
	var updatedFavorites []types.LibraryFavorite
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
func CreateEmptyJSONFile(filePath string) bool {
	emptyData := []byte("[]")
	if err := os.WriteFile(filePath, emptyData, 0644); err != nil {
		return false
	}
	return true
}
