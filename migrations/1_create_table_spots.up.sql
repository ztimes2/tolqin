CREATE TABLE spots (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
    latitude NUMERIC NOT NULL,
	longitude NUMERIC NOT NULL,
	locality TEXT,
	country_code TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);