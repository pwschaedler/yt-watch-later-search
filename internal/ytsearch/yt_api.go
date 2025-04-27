package ytsearch

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func GetYouTubeService() *youtube.Service {
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

	service, err := youtube.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
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

	videos := []Video{}
	for _, item := range playlistItems {
		video, err := parsePlaylistVideoInfo(item)
		if err != nil {
			log.Printf("Skipping, %v\n", err)
			continue
		}
		videos = append(videos, video)
	}
	return videos
}

func GetVideoDetails(service *youtube.Service, videoIds []string) {
	// todo
}

func requestPlaylistChunk(
	service *youtube.Service,
	playlistId string,
	nextPageToken string,
) ([]*youtube.PlaylistItem, string) {
	videoCall := service.PlaylistItems.List([]string{"contentDetails,snippet"})
	videoCall = videoCall.MaxResults(50)
	videoCall = videoCall.PlaylistId(playlistId)
	videoCall = videoCall.PageToken(nextPageToken)
	resp, err := videoCall.Do()
	if err != nil {
		log.Fatalln("Couldn't get videos from playlist.")
	}
	return resp.Items, resp.NextPageToken
}

func requestVideoChunk(service *youtube.Service, videoIds []string, nextPageToken string) ([]*youtube.Video, string) {
	videoCall := service.Videos.List([]string{"contentDetails"})
	videoCall = videoCall.MaxResults(50)
	videoCall = videoCall.Id(strings.Join(videoIds, ","))
	videoCall = videoCall.PageToken(nextPageToken)
	resp, err := videoCall.Do()
	if err != nil {
		log.Fatalln("Couldn't get video details.")
	}
	return resp.Items, resp.NextPageToken
}

func parsePlaylistVideoInfo(item *youtube.PlaylistItem) (Video, error) {
	var err error
	// if item.Status.PrivacyStatus == "private" {
	// 	err = fmt.Errorf("skipping privated video")
	// 	return Video{}, err
	// }

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
		DatePublished: datePublished,
		DateAdded:     dateAdded,
	}, nil
}
