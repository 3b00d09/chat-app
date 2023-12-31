package database

import "database/sql"

func RunSchema(db *sql.DB) {

	const create string = `
	CREATE TABLE IF NOT EXISTS user (
		id TEXT NOT NULL PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS user_session (
		id TEXT NOT NULL PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES user(id),
		active_expires INTEGER NOT NULL,
		idle_expires INTEGER NOT NULL
	);
	`

	db.Exec(create)

}
