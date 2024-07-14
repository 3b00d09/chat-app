package database

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/joho/godotenv"
)
type UserWithWebsocketKeys struct{
	Username string
	WebsocketKey string
}

func init() {
	
	if err := godotenv.Load(); err != nil {
		fmt.Print("Failed to load .env")
		return
	}

}

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

func FetchUsersWithWebsocketKeys(excludedUser string) []UserWithWebsocketKeys {
	statement, err := DB.Prepare("SELECT username, websocket_key FROM user WHERE username != ?")

	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	rows, err := statement.Query(excludedUser)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()


	var results []UserWithWebsocketKeys
	for rows.Next() {
        var user UserWithWebsocketKeys
        err := rows.Scan(&user.Username, &user.WebsocketKey)
        if err != nil {
            log.Fatal(err)
        }
        results = append(results, user)
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

func GenerateCommonWebsocketKey(key1 string, key2 string) string{
    // Sort the keys incase they are in different order
    keys := []string{key1, key2}
    sort.Strings(keys)

    // sum up the keys as they are all integers
    combined := keys[0] + keys[1]

	// grab the salt from env
	salt := os.Getenv("HASH_SALT")

	// decode the salt
	decodedSalt, err := base64.StdEncoding.DecodeString(salt)

	if err != nil {
		return ""
	}

    // hash the combined keys
    h := hmac.New(sha256.New, []byte(decodedSalt))

    // write the combined keys to the HMAC
    h.Write([]byte(combined))

    // get the result and encode as hexadecimal
    result := h.Sum(nil)
    return hex.EncodeToString(result)
}