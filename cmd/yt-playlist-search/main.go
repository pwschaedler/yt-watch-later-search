package main

import (
	"fmt"

	"github.com/pwschaedler/yt-watch-later-search/internal/ytsearch"
	"google.golang.org/api/youtube/v3"
)

func main() {
	youtubeService := ytsearch.GetYouTubeService()

	playlistId := "PLMsroFQMqFwI082s1_mf4oU9ILn-09PfM"
	fmt.Printf("Defaulting to playlist ID: %v\n", playlistId)

	// Getting PublishedAt date doesn't work, that gives the date the video was
	// put into the playlist. Would have to request videos themselves to get
	// actual video date. Probably pull list of video IDs, do a big batch
	// request to get video info for all, and then pull info from there.
	// Or use "contentDetails" instead of "snippet" and get videoPublishedAt
	// from there.

	videos := []ytsearch.Video{}
	var nextPageToken string

	// videoCall := youtubeService.PlaylistItems.List([]string{"snippet"})
	// videoCall = videoCall.MaxResults(50)
	// videoCall = videoCall.PlaylistId(playlistId)
	// resp, err := videoCall.Do()
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// total += len(resp.Items)
	// fmt.Printf("Read %v items, next page token = %v\n", total, resp.NextPageToken)

	// for {
	// 	nextPageToken = resp.NextPageToken
	// 	videoCall.PageToken(nextPageToken)
	// 	resp, err = videoCall.Do()
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}
	// 	total += len(resp.Items)
	// 	fmt.Printf("Read %v items, next page token = %v\n", total, resp.NextPageToken)
	// }

	playlistItems := []*youtube.PlaylistItem{}
	resp, nextPageToken := ytsearch.RequestPlaylistChunk(youtubeService, playlistId, "")
	playlistItems = append(playlistItems, resp...)
	for nextPageToken != "" {
		resp, nextPageToken = ytsearch.RequestPlaylistChunk(youtubeService, playlistId, nextPageToken)
		playlistItems = append(playlistItems, resp...)
	}

	for _, item := range playlistItems {
		video, err := ytsearch.ParsePlaylistVideoInfo(item)
		if err != nil {
			fmt.Printf("Skipping, %v", err)
			continue
		}
		videos = append(videos, video)
	}

	fmt.Printf("Processed %v videos\n", len(videos))
	for i, video := range videos {
		fmt.Printf("%v [%v] %v\n", i, video.Channel, video.Title)
	}

	// fmt.Println(videos)

	// totalResults += len(resp.Items)
	// fmt.Println()
	// fmt.Printf("Read %v items total\n", totalResults)

	// nextPageToken := resp.NextPageToken
	// fmt.Printf("nextPageToken: %v\n", nextPageToken)
	// fmt.Printf("nextPageToken is null: %v\n", nextPageToken == "")
	// fmt.Println()

	// for nextPageToken != "" {
	// 	fmt.Println("Sleeping 5 sec ...")
	// 	time.Sleep(5 * time.Second)

	// 	videoCall.PageToken(nextPageToken)
	// 	resp, err = videoCall.Do()
	// 	ytsearch.HandleError(err, "Couldn't get videos from playlist.")
	// 	for _, item := range resp.Items {
	// 		// t, _ = time.Parse(time.RFC3339, item.Snippet.PublishedAt)
	// 		fmt.Printf("[%v]  %v\n", item.Snippet.VideoOwnerChannelTitle, item.Snippet.Title)
	// 	}
	// 	totalResults += len(resp.Items)
	// 	fmt.Println()
	// 	fmt.Printf("Read %v items total\n", totalResults)
	// 	nextPageToken = resp.NextPageToken
	// 	fmt.Printf("nextPageToken: %v\n", nextPageToken)
	// 	fmt.Printf("nextPageToken is null: %v\n", nextPageToken == "")
	// 	fmt.Println()
	// }

	// fmt.Println("Done!")
}
