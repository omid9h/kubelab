package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"encoding/json"

	_ "github.com/lib/pq"
)

type Book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

var db *sql.DB

// Initialize the database connection
func initDB() {
	var err error
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging the database: %v", err)
	}

	createTable()
}

// Create the books table if it doesn't exist
func createTable() {
	query := `CREATE TABLE IF NOT EXISTS books (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		UNIQUE(title, author)
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Error creating books table: %v", err)
	}
}

// Insert books into the database (if they don't already exist)
func insertBooks(books []Book) {
	for _, book := range books {
		query := `INSERT INTO books (title, author) 
				  VALUES ($1, $2) ON CONFLICT DO NOTHING;`

		_, err := db.Exec(query, book.Title, book.Author)
		if err != nil {
			log.Printf("Error inserting book (%s by %s): %v", book.Title, book.Author, err)
		}
	}
}

// Handler to list all books
func listBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT title, author FROM books;")
	if err != nil {
		http.Error(w, "Error fetching books", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.Title, &book.Author)
		if err != nil {
			http.Error(w, "Error reading book data", http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func main() {
	// Initialize the database
	initDB()

	// Prepopulate some book data
	books := []Book{
		{Title: "The Catcher in the Rye", Author: "J.D. Salinger"},
		{Title: "To Kill a Mockingbird", Author: "Harper Lee"},
		{Title: "1984", Author: "George Orwell"},
	}

	insertBooks(books)

	// Set up the HTTP server
	http.HandleFunc("/books", listBooks)

	log.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
