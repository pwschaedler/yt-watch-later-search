# YouTube Playlist Search

A CLI tool to search and group YouTube videos from a playlist.

Requires a `client_secret.json` to use the YouTube Data API.

```sh
# Run the program while developing
go run ./cmd/yt-playlist-search

# Compile the program for distribution
go build -o bin/
./bin/yt-playlist-search

# Install the program to your Go bin path
go install

# Run tests
go test
```

If the auth token doesn't seem to be working, try the following.

```sh
rm ~/.credentials/youtube-go-quickstart.json
```

## Planned Features

* Search by title, channel, or content
* Sort videos by publish date, date added to playlist, or video length
* Tag videos with similar content
* Cache videos pulled and only pull videos newly added to playlist
* Allow defaulting to commonly used playlist with environment variable/configuration, or offer CLI option for last-used playlist
