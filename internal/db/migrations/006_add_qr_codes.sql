CREATE TABLE IF NOT EXISTS qr_codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    short_url TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(link_id)
);

CREATE INDEX IF NOT EXISTS idx_qr_codes_link_id ON qr_codes(link_id);

ALTER TABLE link_clicks ADD COLUMN qr_code_id INTEGER REFERENCES qr_codes(id);
