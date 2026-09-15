-- Rename existing migration file
-- This migration adds link status and timestamp fields

-- Drop the is_disabled column if it exists (from old migration)
-- Note: SQLite doesn't support DROP COLUMN easily, so we'll just use status instead

ALTER TABLE links ADD COLUMN status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'expired'));
ALTER TABLE links ADD COLUMN disabled_at DATETIME;
ALTER TABLE links ADD COLUMN enabled_at DATETIME;

CREATE INDEX idx_links_status ON links(status);
CREATE INDEX idx_links_disabled_at ON links(disabled_at);
CREATE INDEX idx_links_enabled_at ON links(enabled_at);