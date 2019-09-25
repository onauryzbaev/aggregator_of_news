package main

import (
	"fmt"
	"github.com/gaus57/news-agg/parser"
	"github.com/gaus57/news-agg/repository"
)

func main() {
	p := parser.NewParser()
	news, err := p.Parse(repository.Site{Url: "https://lenta.ru/rss/news", IsRss: true})
	if err != nil {
		panic(err)
	}

	fmt.Println("news", news)
}
