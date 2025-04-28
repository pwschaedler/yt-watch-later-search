package ytsearch

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/adrg/xdg"
	"golang.org/x/oauth2"
	_ "modernc.org/sqlite"
)

func cachePath() string {
	cachePathName := "ytps/cache.db"
	cacheFilePath, err := xdg.CacheFile(cachePathName)
	if err != nil {
		log.Fatalf("Could not get cache file path at %v\n", cachePathName)
	}
	return cacheFilePath
}

func DbConn() {
	cachePath := cachePath()
	db, err := sql.Open("sqlite", cachePath)
	if err != nil {
		log.Fatalf("Could not open cache at %v\n", cachePath)
	}
	log.Printf("Cached at: %v\n", cachePath)
	defer db.Close()

	stmt := `
		CREATE TABLE IF NOT EXISTS tokens (
			created_at INTEGER NOT NULL PRIMARY KEY,
			token_struct TEXT NOT NULL
		)
	`
	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Table 'tokens' created successfully")
}

func WriteToken(token *oauth2.Token) {
	cachePath := cachePath()
	db, err := sql.Open("sqlite", cachePath)
	if err != nil {
		log.Fatalf("Could not open cache at %v\n", cachePath)
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO tokens(created_at, token_struct) VALUES (?, ?)")
	if err != nil {
		log.Fatal("Preparing SQL statement failed")
	}
	defer stmt.Close()

	createdAt := time.Now().Unix()
	tokenJson, err := json.Marshal(token)
	if err != nil {
		log.Fatal("JSON marshalling OAuth token failed")
	}

	_, err = stmt.Exec(createdAt, tokenJson)
	if err != nil {
		log.Fatal("Writing token to database failed")
	}
}

func ReadToken() *oauth2.Token {
	cachePath := cachePath()
	db, err := sql.Open("sqlite", cachePath)
	if err != nil {
		log.Fatalf("Could not open cache at %v\n", cachePath)
	}
	defer db.Close()

	stmt := "SELECT token_struct FROM tokens ORDER BY created_at LIMIT 1"
	rows := db.QueryRow(stmt)
	var tokenJson string
	err = rows.Scan(&tokenJson)
	if err != nil {
		log.Fatal("oh no token read wrong")
	}

	var token oauth2.Token
	err = json.Unmarshal([]byte(tokenJson), &token)
	if err != nil {
		log.Fatal("Unmarshal failed")
	}

	return &token
}
