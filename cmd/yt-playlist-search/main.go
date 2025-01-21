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

	log.Println("Getting YouTube service ...")
	youtubeService := ytsearch.GetYouTubeService()
	log.Println("Authenticated.")
	log.Printf("Getting videos from playlist ID: %v\n", args.PlaylistId)
	videos := ytsearch.GetPlaylistVideos(youtubeService, args.PlaylistId)

	// Example: Search for RuneScape videos, sorted by publish date ascending
	log.Printf("Searching for videos with query: %v\n", args.Query)
	rsVideos := []ytsearch.Video{}

	for _, video := range videos {
		if strings.Contains(strings.ToLower(video.Title), args.Query) ||
			strings.Contains(strings.ToLower(video.Description), args.Query) {
			rsVideos = append(rsVideos, video)
		}
	}

	sort.Slice(rsVideos, func(i, j int) bool {
		return rsVideos[i].DatePublished.Before(rsVideos[j].DatePublished)
	})

	writeResults(&rsVideos)
	log.Printf("Results: %v\n", len(rsVideos))
}

func writeResults(videos *[]ytsearch.Video) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%v\t%v\t%v\t%v\n",
		"PUBLISHED", "CHANNEL", "TITLE", "LINK")
	for _, video := range *videos {
		fmt.Fprintf(w, "%v\t%v\t%v\thttps://www.youtube.com/watch?v=%v\n",
			video.DatePublished.Format(time.DateOnly),
			video.Channel,
			video.Title,
			video.VideoId)
	}
	w.Flush()

}
