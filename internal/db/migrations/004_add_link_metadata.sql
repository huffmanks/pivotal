ALTER TABLE links ADD COLUMN title TEXT;
ALTER TABLE links ADD COLUMN redirect_type TEXT NOT NULL DEFAULT '302' CHECK (redirect_type IN ('301', '302'));
ALTER TABLE links ADD COLUMN fallback_url TEXT;
