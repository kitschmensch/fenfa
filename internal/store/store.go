package store

import (
	"database/sql"
	"fenfa/internal/config"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var dbPath string

type Entry struct {
	Expiration int64  `json:"expiration"`
	Path       string `json:"path"`
}

func Initialize() {
	dbPath = config.BinaryDirectory + `/data.db`
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		db, err := openDB()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		createEntriesSQL := `CREATE TABLE entries (
			hash TEXT PRIMARY KEY,
			expiration INTEGER,
			path TEXT
		);`
		if err := executeSQL(db, createEntriesSQL); err != nil {
			log.Fatal(err)
		}

		createIPAttemptsSQL := `CREATE TABLE ip_attempts (
			ip_address TEXT PRIMARY KEY,
			failed_attempts INTEGER DEFAULT 0
		);`
		if err := executeSQL(db, createIPAttemptsSQL); err != nil {
			log.Fatal(err)
		}
	}
}

func openDB() (*sql.DB, error) {
	return sql.Open("sqlite3", dbPath)
}

func executeSQL(db *sql.DB, sqlStatement string, args ...interface{}) error {
	_, err := db.Exec(sqlStatement, args...)
	return err
}

func Get(hash string) (record Entry, active bool, exists bool) {
	db, err := openDB()
	if err != nil {
		fmt.Println("Error opening database:", err)
		return Entry{}, false, false
	}
	defer db.Close()

	var entry Entry

	query := `SELECT expiration, path FROM entries WHERE hash = ?`
	err = db.QueryRow(query, hash).Scan(&entry.Expiration, &entry.Path)

	if err == sql.ErrNoRows {
		return Entry{}, false, false
	} else if err != nil {
		fmt.Println("Error querying entry:", err)
		return Entry{}, false, false
	}

	currentTime := time.Now().Unix()
	if entry.Expiration <= currentTime {
		return entry, false, true
	}

	return entry, true, true
}

func Add(hash string, expiration int64, path string) error {
	db, err := openDB()
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT INTO entries (hash, expiration, path) VALUES (?, ?, ?) 
		ON CONFLICT(hash) DO UPDATE SET expiration = excluded.expiration, path = excluded.path;`, hash, expiration, path)

	if err != nil {
		return fmt.Errorf("error inserting/updating entry: %v", err)
	}

	return nil
}

func Delete(hash string) error {
	db, err := openDB()
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	return executeSQL(db, `DELETE FROM entries WHERE hash = ?`, hash)
}

func List(table string) error {
	db, err := openDB()
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	query := fmt.Sprintf(`SELECT * FROM %s`, table)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying table %s: %v", table, err)
	}
	defer rows.Close()

	if table == "entries" {
		fmt.Println("Active Share Links:")
		for rows.Next() {
			var hash string
			var entry Entry
			if err := rows.Scan(&hash, &entry.Expiration, &entry.Path); err != nil {
				return fmt.Errorf("error scanning entry: %v", err)
			}
			fmt.Printf("Path: %s, Expiration: %d, Hash: %s\n", entry.Path, entry.Expiration, hash)
		}
	} else if table == "ip_attempts" {
		fmt.Println("IP Attempt Records:")
		for rows.Next() {
			var ipAddress string
			var failedAttempts int
			if err := rows.Scan(&ipAddress, &failedAttempts); err != nil {
				return fmt.Errorf("error scanning IP attempts: %v", err)
			}
			fmt.Printf("IP Address: %s, Failed Attempts: %d\n", ipAddress, failedAttempts)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error during row iteration: %v", err)
	}
	return nil
}

func IncrementFailedAttempts(ip string) error {
	db, err := openDB()
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT INTO ip_attempts (ip_address, failed_attempts) VALUES (?, 1)
		ON CONFLICT(ip_address) DO UPDATE SET failed_attempts = failed_attempts + 1;`, ip)

	if err != nil {
		return fmt.Errorf("error incrementing failed attempts: %v", err)
	}

	return nil
}

func GetFailedAttempts(ip string) (int, error) {
	db, err := openDB()
	if err != nil {
		return 0, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	var failedAttempts int
	query := `SELECT failed_attempts FROM ip_attempts WHERE ip_address = ?`
	err = db.QueryRow(query, ip).Scan(&failedAttempts)

	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("error querying failed attempts: %v", err)
	}

	return failedAttempts, nil
}

func ResetFailedAttempts(ip string) error {
	db, err := openDB()
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	return executeSQL(db, `DELETE FROM ip_attempts WHERE ip_address = ?`, ip)
}
