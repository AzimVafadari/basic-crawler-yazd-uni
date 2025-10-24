package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// The schema for our database. This SQL will be executed if the table doesn't exist.
const schema = `
CREATE TABLE IF NOT EXISTS pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    html_content TEXT,
    status_code INTEGER,
    crawled_at DATETIME NOT NULL
);
`

// Repository holds the database connection pool.
type Repository struct {
	db *sql.DB
}

// NewRepository creates and returns a new Repository. It also ensures the
// database schema is created.
func NewRepository(dbPath string) (*Repository, error) {
	// Open the SQLite database file. It will be created if it doesn't exist.
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	// Ping the database to verify the connection is alive.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	// Execute the schema to create our `pages` table if it doesn't exist.
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("could not create schema: %w", err)
	}

	log.Println("Database connection established and schema verified.")
	return &Repository{db: db}, nil
}

// Close gracefully closes the database connection.
func (r *Repository) Close() {
	if r.db != nil {
		r.db.Close()
		log.Println("Database connection closed.")
	}
}

// SavePage saves the details of a crawled page to the database.
func (r *Repository) SavePage(url string, htmlContent string, statusCode int) error {
	stmt, err := r.db.Prepare("INSERT INTO pages(url, html_content, status_code, crawled_at) VALUES(?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("could not prepare save statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(url, htmlContent, statusCode, time.Now())
	if err != nil {
		return fmt.Errorf("could not execute save statement for url %s: %w", url, err)
	}

	return nil
}

// GetAllURLs retrieves a list of all URLs currently in the database.
func (r *Repository) GetAllURLs() ([]string, error) {
	rows, err := r.db.Query("SELECT url FROM pages")
	if err != nil {
		return nil, fmt.Errorf("could not query for URLs: %w", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, fmt.Errorf("could not scan URL row: %w", err)
		}
		urls = append(urls, url)
	}
	return urls, nil
}
