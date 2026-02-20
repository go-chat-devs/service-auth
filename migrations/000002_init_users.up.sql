CREATE TYPE auth.TWO_FA_TYPE as ENUM ('disable', 'totp');

CREATE TABLE IF NOT EXISTS auth.users (
	id 						SERIAL PRIMARY KEY,
	email 				TEXT NOT NULL UNIQUE,
	password_hash	TEXT NOT NULL,
	two_fa_type		auth.TWO_FA_TYPE NOT NULL DEFAULT 'disable'
);