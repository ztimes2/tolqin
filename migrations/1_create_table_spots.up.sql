CREATE TABLE spots (
	id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	name TEXT NOT NULL,
	latitude NUMERIC NOT NULL,
	longitude NUMERIC NOT NULL,
	locality TEXT,
	country_code TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
