package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const PORT string = ":4040"

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "root"
	dbname   = "db_comsos_blog"
)

type Post struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
var db, err = sql.Open("postgres", psqlInfo)

func setupDatabase() {

	if err != nil {
		panic(err)
	}

	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}

func starServer() {
	http.HandleFunc("/posts", postsHandler)

	fmt.Println("Server is running at http://localhost:" + PORT)

	err := http.ListenAndServe(PORT, nil)

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	setupDatabase()
	starServer()
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getPosts(w, r)
	case "POST":
		createPost(w, r)
	}
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select * from Posts")

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	posts := []Post{}

	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.ID, &p.Title, &p.Content)

		if err != nil {
			fmt.Println(err)
			continue
		}

		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	var post Post

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &post); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("insert into Posts (title, content) values ($1, $2)", post.Title, post.Content)

	if err != nil {
		panic(err)
	}

	lastResult, _ := result.RowsAffected()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lastResult)
}
