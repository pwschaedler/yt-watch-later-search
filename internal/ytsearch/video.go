package ytsearch

import "time"

type Video struct {
	VideoId       string
	Title         string
	Channel       string
	Description   string
	DatePublished time.Time
	DateAdded     time.Time
}
