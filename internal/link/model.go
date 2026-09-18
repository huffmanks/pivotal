package link

import "time"

type LinkStatus string

const (
	LinkStatusActive   LinkStatus = "active"
	LinkStatusDisabled LinkStatus = "disabled"
	LinkStatusExpired  LinkStatus = "expired"
)

type Link struct {
	ID             int64      `json:"id" db:"id"`
	Slug           string     `json:"slug" db:"slug"`
	DestinationURL string     `json:"destination_url" db:"destination_url"`
	Title          string     `json:"title" db:"title"`
	IsCustom       bool       `json:"is_custom" db:"is_custom"`
	ClickCount     int64      `json:"click_count" db:"click_count"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Status         LinkStatus `json:"status" db:"status"`
	DisabledAt     *time.Time `json:"disabled_at,omitempty" db:"disabled_at"`
	EnabledAt      *time.Time `json:"enabled_at,omitempty" db:"enabled_at"`
	RedirectType   string     `json:"redirect_type" db:"redirect_type"`
	FallbackURL    string     `json:"fallback_url,omitempty" db:"fallback_url"`
}

type QRCode struct {
	ID        int64     `json:"id" db:"id"`
	LinkID    int64     `json:"link_id" db:"link_id"`
	ShortURL  string    `json:"short_url" db:"short_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type QRCodeResponse struct {
	ID        int64     `json:"id"`
	ShortURL  string    `json:"short_url"`
	CreatedAt time.Time `json:"created_at"`
}

type AggregatedClick struct {
	Date    string `json:"date,omitempty"`
	Browser string `json:"browser,omitempty"`
	OS      string `json:"os,omitempty"`
	Country string `json:"country,omitempty"`
	Region  string `json:"region,omitempty"`
	City    string `json:"city,omitempty"`
	QRScan  bool   `json:"qr_scan,omitempty"`
	Count   int64  `json:"count"`
}

type ClickEvent struct {
	ID        int64             `json:"id"`
	LinkID    int64             `json:"link_id"`
	QRCodeID  *int64            `json:"-"`
	Referer   string            `json:"referer"`
	UserAgent string            `json:"user_agent"`
	ClickedAt time.Time         `json:"clicked_at"`
	Browser   string            `json:"browser"`
	OS        string            `json:"os"`
	Device    string            `json:"device"`
	Country   string            `json:"country"`
	Region    string            `json:"region"`
	City      string            `json:"city"`
	UTMParams map[string]string `json:"utm_params"`
	QRScan    bool              `json:"qr_scan"`
	IP        string            `json:"ip,omitempty"`
}

type CreateLinkRequest struct {
	Slug           string      `json:"slug,omitempty"`
	DestinationURL string      `json:"destination_url"`
	Title          string      `json:"title,omitempty"`
	ExpiresAt      *time.Time  `json:"expires_at,omitempty"`
	RedirectType   *string     `json:"redirect_type,omitempty"`
	FallbackURL    *string     `json:"fallback_url,omitempty"`
	Status         *LinkStatus `json:"status,omitempty"`
}

type UpdateLinkRequest struct {
	Slug           *string     `json:"slug,omitempty"`
	DestinationURL *string     `json:"destination_url,omitempty"`
	Title          *string     `json:"title,omitempty"`
	ExpiresAt      *time.Time  `json:"expires_at,omitempty"`
	RedirectType   *string     `json:"redirect_type,omitempty"`
	FallbackURL    *string     `json:"fallback_url,omitempty"`
	Status         *LinkStatus `json:"status,omitempty"`
}

type DisableLinkRequest struct {
	FallbackURL *string `json:"fallback_url,omitempty"`
}

type EnableLinkRequest struct{}
