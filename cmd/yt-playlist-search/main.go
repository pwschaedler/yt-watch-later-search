package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/pwschaedler/yt-watch-later-search/internal/ytsearch"
)

func main() {
	// Disable logging if not in verbose mode
	if verbose := os.Getenv("YTPS_VERBOSE"); verbose == "" {
		log.SetOutput(io.Discard)
	}
	args := parseArgs()

	log.Println("Testing database connection ...")
	ytsearch.DbConn()

	log.Println("Getting YouTube service ...")
	youtubeService := ytsearch.GetYouTubeService()
	log.Println("Authenticated.")
	log.Printf("Getting videos from playlist ID: %v\n", args.PlaylistId)
	playlistVideos := ytsearch.GetPlaylistVideos(youtubeService, args.PlaylistId)

	log.Printf("Searching for videos with query: %v\n", args.Query)
	videos := []ytsearch.Video{}

	for _, video := range playlistVideos {
		if strings.Contains(strings.ToLower(video.Title), args.Query) ||
			strings.Contains(strings.ToLower(video.Description), args.Query) {
			videos = append(videos, video)
		}
	}

	sort.Slice(videos, func(i, j int) bool {
		// return videos[i].DatePublished.Before(videos[j].DatePublished)
		return videos[i].Duration < videos[j].Duration
	})

	writeResults(&videos)
	log.Printf("Results: %v\n", len(videos))
}

func writeResults(videos *[]ytsearch.Video) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, '\t', 0)
	fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n",
		"PUBLISHED", "CHANNEL", "TITLE", "DURATION", "LINK")
	for _, video := range *videos {
		link := fmt.Sprintf("https://www.youtube.com/watch?v=%v", video.VideoId)
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n",
			video.DatePublished.Format(time.DateOnly),
			video.Channel,
			video.Title,
			video.Duration,
			link)
	}
	w.Flush()
}
