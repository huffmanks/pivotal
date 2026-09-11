package link

import "time"

type Link struct {
	ID             int64      `json:"id"`
	Slug           string     `json:"slug"`
	DestinationURL string     `json:"destination_url"`
	IsCustom       bool       `json:"is_custom"`
	ClickCount     int64      `json:"click_count"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type ClickEvent struct {
	ID        int64     `json:"id,omitempty"`
	LinkID    int64     `json:"link_id"`
	Referer   string    `json:"referer"`
	UserAgent string    `json:"user_agent"`
	ClickedAt time.Time `json:"clicked_at"`
}
