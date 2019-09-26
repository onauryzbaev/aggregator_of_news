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

	agg := NewAggregator(
		repository.NewRepository(db),
		parser.NewParser(&http.Client{}),
		log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
	)
	agg.Serve()
	defer agg.Stop()

	//news, err := p.Parse(repository.Site{Url: "https://news.rambler.ru/rss/world/", IsRss: true})
	//if err != nil {
	//	panic(err)
	//}

	//news, err := p.Parse(repository.Site{
	//	Url:             "https://www.e1.ru/news/",
	//	IsRss:           false,
	//	NewsItemPath:    "article.e1-article_news",
	//	TitlePath:       ".e1-article__tit",
	//	DescriptionPath: "",
	//	LinkPath:        ".e1-article__link",
	//	DatePath:        ".e1-article__date-text",
	//	ImagePath:       ".e1-article__img",
	//})
	//if err != nil {
	//	panic(err)
	//}
	//
	//j, _ := json.Marshal(news)
	//fmt.Println("news", string(j))
}
