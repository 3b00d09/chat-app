package database

import (
	"crypto/rand"
	"encoding/hex"
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

func GenerateWebSocketKey() (string, error){
	// Generate a random 16 byte key
	bytes := make([]byte, 16)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}