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
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)
type UserWithWebsocketKeys struct{
	Username string
	WebsocketKey string
}

type SidebarUser struct{
	Username string
	Message string
	Created_at string
	WebsocketKey string
}

type SideBarQueryData struct{
	user1_id string
	username1 string
	user1WebsocketKey string
	user2_id string
	username2 string
	user2WebsocketKey string
	message string
	created_at string
}

func init() {
	
	if err := godotenv.Load(); err != nil {
		fmt.Print("Failed to load .env")
		return
	}

}

func FetchSidebarUsers(excludedUser string) []SidebarUser {
	// thanks claude for saving me from this query
	statement, err := DB.Prepare(`
		SELECT
		u1.id AS user1_id,
		u1.username AS user1_username,
		u1.websocket_key AS user1_websocket_key,
		u2.id AS user2_id,
		u2.username AS user2_username,
		u2.websocket_key AS user2_websocket_key,
		m.message AS latest_message,
		m.created_at AS latest_message_time
		FROM
		conversations c
		JOIN user u1 ON c.user1 = u1.id
		JOIN user u2 ON c.user2 = u2.id
		LEFT JOIN messages m ON c.id = m.conversation_id
		LEFT JOIN messages m2 ON c.id = m2.conversation_id
		AND m.created_at < m2.created_at
		WHERE
		(
			c.user1 = ?
			OR c.user2 = ?
		)
		AND m2.id IS NULL
		ORDER BY
		m.created_at DESC;
  `)

	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	rows, err := statement.Query(excludedUser, excludedUser)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var results []SidebarUser

	for rows.Next() {
        var row SideBarQueryData
		var user SidebarUser
        err := rows.Scan(&row.user1_id, &row.username1, &row.user1WebsocketKey, &row.user2_id, &row.username2, &row.user2WebsocketKey, &row.message, &row.created_at)        
		if err != nil {
            if err.Error() == `sql: Scan error on column index 2, name "latest_message": converting NULL to string is unsupported` {
                row.message = ""
				row.created_at = ""
            }else{
				// if the error is not the one we are expecting, ignore the row and go to the next one
				continue
			}
        }else {
			// if the user1_id is the excluded user, then the username and websocket key should be from user2
			// no nicer way of doing this since our excluded user can either be user1 or user2 depending on who initiated the conversation
			if row.user1_id == excludedUser {
				user.Username = row.username2
				user.WebsocketKey = row.user2WebsocketKey
			}else {
				user.Username = row.username1
				user.WebsocketKey = row.user1WebsocketKey
			}
			user.Message = row.message
			user.Created_at = FormatTime(row.created_at)
			results = append(results, user)
			}
		
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

func FetchConversationId(user1 string, user2 string) string{
	statement, err := DB.Prepare("SELECT id FROM conversations WHERE (user1 = ? AND user2 = ?) OR (user1 = ? AND user2 = ?)")

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return ""
		}
	}
	defer statement.Close()

	var id string
	statement.QueryRow(user1, user2, user2, user1).Scan(&id)

	return id
}

func CreateConversation(user1 string, user2 string) string{

	statement, err := DB.Prepare("INSERT INTO conversations (id, user1, user2) VALUES (?, ?, ?)")

	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	uid, _ := uuid.NewRandom()

	rowId := uid.String()

	statement.Exec(rowId, user1, user2)

	return rowId
}

func FormatTime(unixTime string) string{
	// convert the unix timestamp to a time.Time object
	unixTimeInt, err := strconv.ParseInt(unixTime, 10, 64)

	if err != nil {
		return ""
	}

	createdTime := time.Unix(unixTimeInt, 0)
	now := time.Now()
	diff := now.Sub(createdTime)

	var timeAgo string
	switch {
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		timeAgo = strconv.Itoa(minutes) + "m"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		timeAgo = strconv.Itoa(hours) + "h"
	default:
		days := int(diff.Hours() / 24)
		timeAgo = strconv.Itoa(days) + "d"
	}

	return timeAgo
}