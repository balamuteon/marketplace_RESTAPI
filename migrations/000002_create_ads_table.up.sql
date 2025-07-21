CREATE TABLE IF NOT EXISTS ads (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title TEXT NOT NULL CHECK (
		length(title) >= 1
		AND length(title) <= 100
	),
	description TEXT NOT NULL CHECK (length(description) <= 1000),
	price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
	image_url TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ads_user_id ON ads(user_id);