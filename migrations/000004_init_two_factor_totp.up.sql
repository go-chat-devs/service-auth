CREATE TABLE IF NOT EXISTS auth.two_factor_totp (
	user_id	SERIAL PRIMARY KEY,
	secret	VARCHAR(32) NOT NULL
);