package database

type User struct {
	ID       string `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

type UserSession struct {
	ID            string `db:"id"`
	UserID        string `db:"user_id"`
	ActiveExpires int64  `db:"active_expires"`
	IdleExpires   int64  `db:"idle_expires"`
}
