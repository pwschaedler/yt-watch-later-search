package ytsearch

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/sosodev/duration"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func GetYouTubeService() *youtube.Service {
	token := ReadToken()

	if false {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		config := oauth2.Config{
			ClientID:     os.Getenv("YTAPI_CLIENT_ID"),
			ClientSecret: os.Getenv("YTAPI_CLIENT_SECRET"),
			Endpoint:     google.Endpoint,
			RedirectURL:  "http://localhost:8080/callback",
			Scopes:       []string{youtube.YoutubeReadonlyScope},
		}

		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		browser.OpenURL(authURL)

		srv := &http.Server{Addr: ":8080"}
		var authCode string
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			authCode = r.FormValue("code")
			// The browser doesn't allow us to close the tab for the user
			w.Write([]byte("You may now close this tab."))
			cancel()
		})

		go func() {
			srv.ListenAndServe()
		}()
		<-ctx.Done()

		ctx = context.Background()
		token, err := config.Exchange(ctx, authCode)
		if err != nil {
			log.Fatalln("Failed to retrieve auth token.")
		}
		WriteToken(token)
	}

	ctx := context.Background()
	config := oauth2.Config{
		ClientID:     os.Getenv("YTAPI_CLIENT_ID"),
		ClientSecret: os.Getenv("YTAPI_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{youtube.YoutubeReadonlyScope},
	}

	service, err := youtube.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	log.Printf("Updated token: %v\n", token)
	if err != nil {
		log.Fatalln("Failed to create YouTube client.")
	}
	return service
}

func GetPlaylistVideos(service *youtube.Service, playlistId string) []Video {
	var nextPageToken string
	playlistItems := []*youtube.PlaylistItem{}
	resp, nextPageToken := requestPlaylistChunk(service, playlistId, "")
	playlistItems = append(playlistItems, resp...)
	for nextPageToken != "" {
		resp, nextPageToken = requestPlaylistChunk(service, playlistId, nextPageToken)
		playlistItems = append(playlistItems, resp...)
	}

	publicPlaylistItems := []*youtube.PlaylistItem{}
	for _, item := range playlistItems {
		if item.Status.PrivacyStatus == "public" || item.Status.PrivacyStatus == "unlisted" {
			publicPlaylistItems = append(publicPlaylistItems, item)
		}
	}

	videoIds := []string{}
	for _, item := range publicPlaylistItems {
		videoIds = append(videoIds, item.ContentDetails.VideoId)
	}
	fmt.Printf("Video IDs length: %d\n", len(videoIds))
	durations := GetVideoDurations(service, videoIds)

	videos := []Video{}
	fmt.Printf("Playlist items length: %d\n", len(publicPlaylistItems))
	fmt.Printf("Durations length: %d\n", len(durations))
	for idx, item := range publicPlaylistItems {
		d := durations[idx]
		video, err := parsePlaylistVideoInfo(item, d)
		if err != nil {
			log.Printf("Skipping, %v\n", err)
			continue
		}
		videos = append(videos, video)
	}

	return videos
}

// Makes API requests to get details about individual videos that aren't
// available from the PlaylistItems. This API call doesn't support paging or
// responses with more than 50 items, so we have to manually chunk the IDs and
// only make requests with up to 50 IDs.
func GetVideoDurations(service *youtube.Service, videoIds []string) []time.Duration {
	videoCall := service.Videos.List([]string{"contentDetails"})
	videoItems := []*youtube.Video{}
	maxItems := 50 // Enforced by the API
	for videoIdChunk := range slices.Chunk(videoIds, maxItems) {
		videoCall = videoCall.Id(strings.Join(videoIdChunk, ","))
		resp, err := videoCall.Do()
		if err != nil {
			log.Fatalln("Couldn't get video details.")
		}
		videoItems = append(videoItems, resp.Items...)
	}

	durations := []time.Duration{}
	for _, item := range videoItems {
		rawDuration := item.ContentDetails.Duration
		dd, err := duration.Parse(rawDuration)
		if err != nil {
			log.Fatalln("Couldn't parse ISO 8601 duration")
		}
		d := dd.ToTimeDuration()
		durations = append(durations, d)
	}

	return durations
}

func requestPlaylistChunk(
	service *youtube.Service,
	playlistId string,
	nextPageToken string,
) ([]*youtube.PlaylistItem, string) {
	videoCall := service.PlaylistItems.List([]string{"contentDetails,snippet,status"})
	videoCall = videoCall.MaxResults(50)
	videoCall = videoCall.PlaylistId(playlistId)
	videoCall = videoCall.PageToken(nextPageToken)
	resp, err := videoCall.Do()
	if err != nil {
		log.Fatalln("Couldn't get videos from playlist.")
	}
	return resp.Items, resp.NextPageToken
}

func parsePlaylistVideoInfo(item *youtube.PlaylistItem, duration time.Duration) (Video, error) {
	var err error

	title := item.Snippet.Title
	stringPublished := item.ContentDetails.VideoPublishedAt
	stringAdded := item.Snippet.PublishedAt

	datePublished, err := time.Parse(time.RFC3339, stringPublished)
	if err != nil {
		err = fmt.Errorf("cannot parse videoPublishedAt date: %v, %v", title, stringPublished)
		return Video{}, err
	}
	dateAdded, err := time.Parse(time.RFC3339, stringAdded)
	if err != nil {
		err = fmt.Errorf("cannot parse PublishedAt date: %v, %v", title, stringAdded)
		return Video{}, err
	}

	return Video{
		VideoId:       item.ContentDetails.VideoId,
		Title:         title,
		Channel:       item.Snippet.VideoOwnerChannelTitle,
		Description:   item.Snippet.Description,
		Duration:      duration,
		DatePublished: datePublished,
		DateAdded:     dateAdded,
	}, nil
}
