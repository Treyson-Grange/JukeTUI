package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	clientID := os.Getenv("SPOTIFY_ID")
	clientSecret := os.Getenv("SPOTIFY_SECRET")

	fmt.Println("Opening login page...")
	OpenLoginPage(clientID)

	code := GetCodeFromCallback()
	token, err := GetSpotifyToken(context.Background(), clientID, clientSecret, code)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Println("Login successful! Access token retrieved.")
	fmt.Println("Fetching recommendations...")

	//POC, delete this
	fetchRecommendations(token.AccessToken)

	pbState := handleGenericFetch[PlaybackState](PLAYBACK_ENDPOINT, token.AccessToken)
	fmt.Printf("Playback state: %+v\n", pbState.Device.ID)
	// fmt.Println(testPut("https://api.spotify.com/v1/me/player/play", token.AccessToken, pbState.Device.ID))
	// fmt.Println(testPut("https://api.spotify.com/v1/me/player/pause", token.AccessToken, pbState.Device.ID))
}

func handleGenericFetch[T any](endpoint string, accessToken string) T {
	data, err := genericFetch[T](endpoint, accessToken)
	if err != nil {
		log.Fatalf("Failed to fetch playback state: %v", err)
	}
	return data
}

// POC, delete this
func fetchRecommendations(accessToken string) {
	// Example: Use the Spotify API to fetch recommendations.
	recs, err := GetRecommendations(accessToken, "pop", 5)
	if err != nil {
		log.Fatalf("Failed to fetch recommendations: %v", err)
	}

	fmt.Println("Recommended Tracks:")
	for _, track := range recs.Tracks {
		fmt.Printf("- %s by %s\n", track.Name, track.Artists[0].Name)
	}
}
