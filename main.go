package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gaus57/news-agg/parser"
	"github.com/gaus57/news-agg/repository"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	interval := flag.Int("i", 600, "Parsing interval in seconds")
	port := flag.Int("p", 8080, "Port for http server")
	flag.Parse()

	db, err := gorm.Open("postgres", "host=localhost port=54320 user=postgres dbname=newsagg sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}
	defer db.Close()

	app := NewApplication(
		repository.NewRepository(db),
		parser.NewParser(&http.Client{}),
		log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		*port,
		time.Duration(*interval)*time.Second,
	)
	app.Serve()
	defer app.Stop()
}
