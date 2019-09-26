package main

import (
	"encoding/json"
	"fmt"
	"github.com/gaus57/news-agg/parser"
	"github.com/gaus57/news-agg/repository"
)

func main() {
	p := parser.NewParser()
	//news, err := p.Parse(repository.Site{Url: "https://news.rambler.ru/rss/world/", IsRss: true})
	//if err != nil {
	//	panic(err)
	//}

	news, err := p.Parse(repository.Site{
		Url:             "https://www.e1.ru/news/",
		IsRss:           false,
		NewsItemPath:    "article.e1-article_news",
		TitlePath:       ".e1-article__tit",
		DescriptionPath: "",
		LinkPath:        ".e1-article__link",
		DatePath:        ".e1-article__date-text",
		ImagePath:       ".e1-article__img",
	})
	if err != nil {
		panic(err)
	}

	j, _ := json.Marshal(news)
	fmt.Println("news", string(j))
}
