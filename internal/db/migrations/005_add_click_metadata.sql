ALTER TABLE link_clicks ADD COLUMN browser TEXT;
ALTER TABLE link_clicks ADD COLUMN os TEXT;
ALTER TABLE link_clicks ADD COLUMN device TEXT;
ALTER TABLE link_clicks ADD COLUMN country TEXT;
ALTER TABLE link_clicks ADD COLUMN region TEXT;
ALTER TABLE link_clicks ADD COLUMN city TEXT;
ALTER TABLE link_clicks ADD COLUMN utm_params TEXT;
ALTER TABLE link_clicks ADD COLUMN qr_scan INTEGER NOT NULL DEFAULT 0;
ALTER TABLE link_clicks ADD COLUMN ip TEXT;
