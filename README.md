# YouTube Playlist Search

A CLI tool to search and group YouTube videos from a playlist.

Requires a Google client secret to use the YouTube Data API. This can be generated from the Google Cloud Console. Create a project to use with this tool. Under "APIs & Services", click "Credentials". Click "Create Credentials", and choose "OAuth client ID". Choose application type "Desktop", and the name doesn't matter. Take the provided Client ID and Client secret and use them with environment variables `YTAPI_CLIENT_ID` and `YTAPI_CLIENT_SECRET`, respectively.

```sh
# Run the program while developing
go run ./cmd/yt-playlist-search

# Compile the program for distribution
go build -o ./bin/ ./cmd/yt-playlist-search
./bin/yt-playlist-search

# Install the program to your Go bin path
go install

# Run tests
go test
```

## Configuration

The following environment variables can be used to configure the tool.

* `YTAPI_CLIENT_ID`: Client ID provided by Google Cloud Console.
* `YTAPI_CLIENT_SECRET`: Client secret provided by Google Cloud Console.
* `YTPS_PLAYLIST_ID`: Default playlist ID to search.
* `YTPS_VERBOSE`: Enables logging when non-empty.

## Planned Features

* Search by title, channel, or content
* Sort videos by publish date, date added to playlist, or video length
* Tag videos with similar content
* Cache videos pulled and only pull videos newly added to playlist
* Allow defaulting to commonly used playlist with environment variable/configuration, or offer CLI option for last-used playlist
