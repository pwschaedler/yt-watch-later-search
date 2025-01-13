package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func main() {
	ctx := context.Background()
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
		srv.Shutdown(ctx)
	})
	srv.ListenAndServe()

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		log.Fatalln("Failed to retrieve auth token.")
	}

	_, err = youtube.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	if err != nil {
		log.Fatalln("Failed to create YouTube client.")
	}
}
