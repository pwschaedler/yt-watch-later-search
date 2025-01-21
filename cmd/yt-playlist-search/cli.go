package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Args struct {
	PlaylistId string
	Query      string
}

func parseArgs() Args {
	args := Args{}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "  query string\n")
		fmt.Fprintf(os.Stderr, "    \tQuery to search for.\n")
	}
	flag.StringVar(&args.PlaylistId, "playlistId", os.Getenv("YTPS_PLAYLIST_ID"), "Playlist ID to search through.")
	flag.Parse()
	queryArgs := flag.Args()
	args.Query = strings.Join(queryArgs, " ")
	return args
}
