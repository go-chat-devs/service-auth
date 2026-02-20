CREATE TABLE IF NOT EXISTS auth.sessions (
	session_key	TEXT PRIMARY KEY,
	user_uid		UUID NOT NULL UNIQUE
);