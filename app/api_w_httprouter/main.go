package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	"github.com/julienschmidt/httprouter"
	"github.com/pobyzaarif/belajarGoB16/util/db"
)

var v = validator.New()

func main() {
	db, err := db.InitDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	defer db.Close()

	router := httprouter.New()

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, i interface{}) {
		fmt.Println(i)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode("panic processing request")
	}

	router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode("hello world")
	})

	router.GET("/panic", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		listAnimal := []string{"chicken", "rabbit", "cat", "dog"}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listAnimal[10])
	})

	router.GET("/students", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		students, err := GetStudents(db)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode("error processing request")
		}

		_ = json.NewEncoder(w).Encode(students)
	})

	router.POST("/students", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")

		// take request
		decoder := json.NewDecoder(r.Body)
		var student Student
		err := decoder.Decode(&student)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode("error payload body")
			return
		}

		// validate body request
		errValidation := v.Struct(student)
		if errValidation != nil {
			fmt.Println(errValidation)
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode("error specification not match")
			return
		}

		// processing request
		err = CreateStudent(db, student.Name, student.Age)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode("error processing request")
			return
		}

		// postive response
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode("Created")
	})

	server := http.Server{
		Addr:    os.Getenv("APP_HOST"),
		Handler: router,
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

type Student struct {
	ID   int    `json:"id"`
	Name string `json:"name" validate:"max=10,min=3"`
	Age  int    `json:"age" validate:"required"`
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

func CreateStudent(h *sql.DB, name string, age int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.ExecContext(ctx, "INSERT INTO students (name, age) VALUES (?, ?)", name, age)
	if err != nil {
		return err
	}

	return nil
}
