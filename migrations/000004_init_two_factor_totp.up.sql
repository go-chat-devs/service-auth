CREATE TABLE IF NOT EXISTS auth.two_factor_totp (
	user_uid	UUID PRIMARY KEY,
	secret		VARCHAR(32) NOT NULL
);