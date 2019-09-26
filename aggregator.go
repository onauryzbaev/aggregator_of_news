package main

import (
	"github.com/gaus57/news-agg/repository"
	"log"
	"net/http"
	"time"
)

type Repository interface {
	Migrate()
	GetSites() ([]repository.Site, error)
	AddSite(site *repository.Site) error
	UpdateSite(site repository.Site) error
	DeleteSite(id int) error
	GetNews(offset int, limit int, search string) ([]repository.NewsItem, error)
	AddNewsItem(item *repository.NewsItem) error
}

type Parser interface {
	Parse(site repository.Site) ([]repository.NewsItem, error)
}

type aggregator struct {
	log        *log.Logger
	repository Repository
	parser     Parser
	stop       chan bool
}

func NewAggregator(repository Repository, parser Parser, loger *log.Logger) *aggregator {
	return &aggregator{
		log:        loger,
		repository: repository,
		parser:     parser,
		stop:       make(chan bool),
	}
}

func (agg *aggregator) Serve() {
	agg.repository.Migrate()
	agg.parsing()
	agg.serveHttp()
}

func (agg *aggregator) Stop() {
	close(agg.stop)
}

func (agg *aggregator) parsing() {
	go func() {
		agg.parseSites()
		ticker := time.NewTicker(time.Minute * 1)
		select {
		case <-agg.stop:
			return
		case <-ticker.C:
			agg.parseSites()
		}
	}()
}

func (agg *aggregator) parseSites() {
	sites, err := agg.repository.GetSites()
	if err != nil {
		agg.log.Printf("Failed get sites from repository: %v", err)

		return
	}

	for _, site := range sites {
		news, err := agg.parser.Parse(site)
		if err != nil {
			agg.log.Printf("Failed parse site %s: %v", site.Url, err)
			continue
		}

		insert := 0
		for _, item := range news {
			item.SiteID = site.ID
			err := agg.repository.AddNewsItem(&item)
			if err != nil {
				agg.log.Printf("Failed add news to repository: %v", err)
				continue
			}
			insert++
		}

		agg.log.Printf("Complete parse site %s - add %d news", site.Url, insert)
	}
}

func (agg *aggregator) serveHttp() {
	http.HandleFunc("/", agg.mainHandler)
	agg.log.Fatal(http.ListenAndServe(":8080", nil))
}

func (agg *aggregator) mainHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(200)
	res.Write([]byte("test"))
}
