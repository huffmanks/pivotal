package link

import "time"

type Link struct {
	ID             int64     `json:"id"`
	Slug           string    `json:"slug"`
	DestinationURL string    `json:"destination_url"`
	IsCustom       bool      `json:"is_custom"`
	ClickCount     int64     `json:"click_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type ClickEvent struct {
	LinkID    int64
	Referer   string
	UserAgent string
	ClickedAt time.Time
}
