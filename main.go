package main

import (
	"fmt"
	"github.com/gaus57/news-agg/parser"
	"github.com/gaus57/news-agg/repository"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"net/http"
	"os"
)

func main() {
	db, err := gorm.Open("postgres", "host=localhost port=54320 user=postgres dbname=newsagg sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}
	defer db.Close()

	app := NewApplication(
		repository.NewRepository(db),
		parser.NewParser(&http.Client{}),
		log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
	)
	app.Serve()
	defer app.Stop()
}
