package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pwschaedler/yt-watch-later-search/internal/ytsearch"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

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
	token := ytsearch.GetToken(ctx, config)
	youtubeService, err := youtube.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	ytsearch.HandleError(err, "Error creating YouTube client")

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
	ytsearch.HandleError(err, "Couldn't get videos from playlist.")
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
		ytsearch.HandleError(err, "Couldn't get videos from playlist.")
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
