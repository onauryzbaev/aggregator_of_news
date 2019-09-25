package main

import "github.com/gaus57/news-agg/repository"

type Repository interface {
	GetSites() ([]repository.Site, error)
	AddSite(site repository.Site) error
	UpdateSite(id int, site repository.Site) error
	GetNews(page int, search string) ([]repository.NewsItem, error)
	AddNewsItem(item repository.NewsItem) error
}

type Parser interface {
	Parse(site repository.Site) ([]repository.NewsItem, error)
}
