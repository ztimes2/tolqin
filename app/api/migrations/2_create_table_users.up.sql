CREATE TYPE user_role AS ENUM ('admin');

CREATE TABLE users (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    role user_role NOT NULL,
    password_hash TEXT NOT NULL,
    password_salt TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);