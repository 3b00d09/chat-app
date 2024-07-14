package database

import "database/sql"

func RunSchema(db *sql.DB) {

	const create string = `
	CREATE TABLE IF NOT EXISTS user (
		id TEXT NOT NULL PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		websocket_key TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS user_session (
		id TEXT NOT NULL PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES user(id),
		active_expires INTEGER NOT NULL,
		idle_expires INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS messages(
		id TEXT NOT NULL PRIMARY KEY,
		sender TEXT NOT NULL REFERENCES user(id),
		recipient TEXT NOT NULL REFERENCES user(id),
		message TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);
	`

	db.Exec(create)

}
