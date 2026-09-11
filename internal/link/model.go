package link

import "time"

type Link struct {
	ID             int64      `db:"id" json:"id"`
	Slug           string     `db:"slug" json:"slug"`
	DestinationURL string     `db:"destination_url" json:"destination_url"`
	IsCustom       bool       `db:"is_custom" json:"is_custom"`
	ClickCount     int64      `db:"click_count" json:"click_count"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	ExpiresAt      *time.Time `db:"expires_at" json:"expires_at"`
}

type ClickEvent struct {
	ID        int64     `db:"-" json:"id"`
	LinkID    int64     `db:"link_id" json:"link_id"`
	Referer   string    `db:"referer" json:"referer"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	ClickedAt time.Time `db:"clicked_at" json:"clicked_at"`
}
