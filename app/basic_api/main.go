package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pobyzaarif/belajarGoB16/util/db"
)

type Employee struct {
	FullName string `json:"full_name"` // commonly json use snake_case style
	Age      int
}

func main() {
	db, err := db.InitDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world")
	})

	mux.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "this is home page")
	})

	mux.HandleFunc("/sample-json", func(w http.ResponseWriter, r *http.Request) {
		// data := map[string]interface{}{
		// 	"name": "pobs",
		// 	"age":  20,
		// }

		var data Employee
		data.FullName = "pobss"
		data.Age = 20

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		students, err := GetStudents(db)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode("error processing request")
		}

		_ = json.NewEncoder(w).Encode(students)
	})

	mux.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		posts, err := RequestGetPostsToJSONPlaceholder()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode("error processing request")
		}
		posts = posts[90:] // assume i onluy need 10 record from post

		_ = json.NewEncoder(w).Encode(posts)
	})

	server := http.Server{
		Addr:    os.Getenv("APP_HOST"),
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

type Student struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func GetStudents(h *sql.DB) ([]Student, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := h.QueryContext(ctx, "SELECT id, name, age FROM students")
	if err != nil {
		fmt.Println("Error querying data:", err)
		return nil, err
	}
	defer rows.Close()

	var students []Student

	for rows.Next() {
		var id int
		var name string
		var age int
		err := rows.Scan(&id, &name, &age)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			return nil, err
		}
		students = append(students, Student{ID: id, Name: name, Age: age})
	}

	err = rows.Err()
	if err != nil {
		fmt.Println("Error with rows:", err)
		return nil, err
	}
	return students, nil
}

func RequestGetPostsToJSONPlaceholder() (response []map[string]interface{}, err error) {
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		return response, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}

	return response, err
}
