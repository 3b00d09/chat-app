package database

import (
	"log"
)

func FetchSidebarUsers(excludedUser string) []string {
	statement, err := DB.Prepare("SELECT username FROM user WHERE username != ?")

	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	rows, err := statement.Query(excludedUser)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
        var username string
        err := rows.Scan(&username)
        if err != nil {
            log.Fatal(err)
        }
        results = append(results, username)
    }

	return results
}
