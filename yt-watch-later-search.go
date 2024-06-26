package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Get auth token.
func getToken(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return tok
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("youtube-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func main() {
	ctx := context.Background()
	key, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(key, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	token := getToken(ctx, config)
	youtubeService, err := youtube.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	handleError(err, "Error creating YouTube client")

	// The real stuff - Get list of playlists
	// call := youtubeService.Playlists.List([]string{"snippet"})
	// call = call.MaxResults(100)
	// call = call.Mine(true)
	// response, err := call.Do()
	// handleError(err, "Couldn't request playlists.")
	// for _, item := range response.Items {
	// 	fmt.Println(item.Snippet.Title)
	// }

	// Get videos from playlist
	// PLMsroFQMqFwI082s1_mf4oU9ILn-09PfM
	fmt.Println("In case you forgot: PLMsroFQMqFwI082s1_mf4oU9ILn-09PfM")
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Playlist ID: ")
	scanner.Scan()
	playlistId := scanner.Text()
	// fmt.Println(playlistId)

	// Getting PublishedAt date doesn't work, that gives the date the video was
	// put into the playlist. Would have to request videos themselves to get
	// actual video date. Probably pull list of video IDs, do a big batch
	// request to get video info for all, and then pull info from there.
	// Or use "contentDetails" instead of "snippet" and get videoPublishedAt
	// from there.

	totalResults := 0

	videoCall := youtubeService.PlaylistItems.List([]string{"snippet"})
	videoCall = videoCall.MaxResults(50)
	videoCall = videoCall.PlaylistId(playlistId)
	resp, err := videoCall.Do()
	handleError(err, "Couldn't get videos from playlist.")
	// var t time.Time
	for _, item := range resp.Items {
		// t, _ = time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		// , t.Format(time.DateOnly)
		// fmt.Printf("Date compare -- %v // %v\n", item.Snippet.PublishedAt, t)
		fmt.Printf("[%v]  %v\n", item.Snippet.VideoOwnerChannelTitle, item.Snippet.Title)
	}

	totalResults += len(resp.Items)
	fmt.Println()
	fmt.Printf("Read %v items total\n", totalResults)

	nextPageToken := resp.NextPageToken
	fmt.Printf("nextPageToken: %v\n", nextPageToken)
	fmt.Printf("nextPageToken is null: %v\n", nextPageToken == "")
	fmt.Println()

	for nextPageToken != "" {
		fmt.Println("Sleeping 5 sec ...")
		time.Sleep(5 * time.Second)

		videoCall.PageToken(nextPageToken)
		resp, err = videoCall.Do()
		handleError(err, "Couldn't get videos from playlist.")
		for _, item := range resp.Items {
			// t, _ = time.Parse(time.RFC3339, item.Snippet.PublishedAt)
			fmt.Printf("[%v]  %v\n", item.Snippet.VideoOwnerChannelTitle, item.Snippet.Title)
		}
		totalResults += len(resp.Items)
		fmt.Println()
		fmt.Printf("Read %v items total\n", totalResults)
		nextPageToken = resp.NextPageToken
		fmt.Printf("nextPageToken: %v\n", nextPageToken)
		fmt.Printf("nextPageToken is null: %v\n", nextPageToken == "")
		fmt.Println()
	}

	fmt.Println("Done!")
}
