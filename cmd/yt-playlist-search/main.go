package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pwschaedler/yt-watch-later-search/internal/ytsearch"
	"google.golang.org/api/youtube/v3"
)

func main() {
	youtubeService := ytsearch.GetYouTubeService()

	playlistId := "PLMsroFQMqFwI082s1_mf4oU9ILn-09PfM"
	fmt.Printf("Defaulting to playlist ID: %v\n", playlistId)

	var nextPageToken string
	playlistItems := []*youtube.PlaylistItem{}
	resp, nextPageToken := ytsearch.RequestPlaylistChunk(youtubeService, playlistId, "")
	playlistItems = append(playlistItems, resp...)
	for nextPageToken != "" {
		resp, nextPageToken = ytsearch.RequestPlaylistChunk(youtubeService, playlistId, nextPageToken)
		playlistItems = append(playlistItems, resp...)
	}

	videos := []ytsearch.Video{}
	for _, item := range playlistItems {
		video, err := ytsearch.ParsePlaylistVideoInfo(item)
		if err != nil {
			fmt.Printf("Skipping, %v\n", err)
			continue
		}
		videos = append(videos, video)
	}

	// Example: Search for RuneScape videos, sorted by publish date ascending
	rsVideos := []ytsearch.Video{}

	for _, video := range videos {
		if strings.Contains(strings.ToLower(video.Title), "runescape") ||
			strings.Contains(strings.ToLower(video.Title), "osrs") ||
			strings.Contains(strings.ToLower(video.Description), "runescape") ||
			strings.Contains(strings.ToLower(video.Description), "osrs") {
			rsVideos = append(rsVideos, video)
		}
	}

	sort.Slice(rsVideos, func(i, j int) bool {
		return rsVideos[i].DatePublished.Before(rsVideos[j].DatePublished)
	})

	fmt.Println("Videos mentioning RuneScape:")
	for _, video := range rsVideos {
		fmt.Printf(
			"%v [%v] %v (https://www.youtube.com/watch?v=%v)\n",
			video.DatePublished.Format(time.DateOnly),
			video.Channel,
			video.Title,
			video.VideoId,
		)
	}
	fmt.Printf("Results: %v\n", len(rsVideos))
}
