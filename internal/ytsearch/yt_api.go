package ytsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Get auth token.
func getToken(config *oauth2.Config) *oauth2.Token {
	cacheFile, err := TokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := TokenFromFile(cacheFile)
	if err != nil {
		tok = GetTokenFromWeb(config)
		SaveToken(cacheFile, tok)
	}
	return tok
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func TokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("youtube-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func SaveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func HandleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func GetYouTubeService() *youtube.Service {
	ctx := context.Background()
	key, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(key, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	token := getToken(config)
	youtubeService, err := youtube.NewService(
		ctx,
		option.WithTokenSource(config.TokenSource(ctx, token)),
	)
	HandleError(err, "Error creating YouTube client")
	return youtubeService
}

func RequestPlaylistChunk(
	service *youtube.Service,
	playlistId string,
	nextPageToken string,
) ([]*youtube.PlaylistItem, string) {
	videoCall := service.PlaylistItems.List([]string{"contentDetails,snippet"})
	videoCall = videoCall.MaxResults(50)
	videoCall = videoCall.PlaylistId(playlistId)
	videoCall = videoCall.PageToken(nextPageToken)
	resp, err := videoCall.Do()
	HandleError(err, "Couldn't get videos from playlist.")
	return resp.Items, resp.NextPageToken
}

func ParsePlaylistVideoInfo(item *youtube.PlaylistItem) (Video, error) {
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
