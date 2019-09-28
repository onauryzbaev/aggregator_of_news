package main

import (
	"errors"
	"github.com/gaus57/news-agg/repository"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

type mockedParser struct {
	mock.Mock
}

func (pars *mockedParser) Parse(site repository.Site) ([]repository.NewsItem, error) {
	args := pars.MethodCalled("Parse", site)

	return args.Get(0).([]repository.NewsItem), args.Error(1)
}

type mockedRepository struct {
	mock.Mock
}

func (rep *mockedRepository) Migrate() {
	rep.MethodCalled("Migrate")
}

func (rep *mockedRepository) GetSites() ([]repository.Site, error) {
	args := rep.MethodCalled("GetSites")

	return args.Get(0).([]repository.Site), args.Error(1)
}

func (rep *mockedRepository) AddSite(site *repository.Site) error {
	args := rep.MethodCalled("AddSite", site)

	return args.Error(0)
}

func (rep *mockedRepository) DeleteSite(id int) error {
	args := rep.MethodCalled("DeleteSite", id)

	return args.Error(0)
}

func (rep *mockedRepository) GetNews(offset int, limit int, search string) ([]repository.NewsItem, error) {
	args := rep.MethodCalled("GetNews", offset, limit, search)

	return args.Get(0).([]repository.NewsItem), args.Error(1)
}

func (rep *mockedRepository) HasNewsItem(item repository.NewsItem) (bool, error) {
	args := rep.MethodCalled("HasNewsItem", item)

	return args.Bool(0), args.Error(1)
}

func (rep *mockedRepository) AddNewsItem(item *repository.NewsItem) error {
	args := rep.MethodCalled("AddNewsItem", item)

	return args.Error(0)
}

func getApplication() *application {
	return &application{
		log:        log.New(ioutil.Discard, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		repository: new(mockedRepository),
		parser:     new(mockedParser),
		stop:       make(chan bool),
	}
}

func TestParsing(t *testing.T) {
	app := getApplication()
	app.interval = time.Millisecond * 200

	site1 := repository.Site{ID: 1, Url: "http://test1.ru", IsRss: true}
	site2 := repository.Site{ID: 2, Url: "http://test2.ru", IsRss: false}
	site3 := repository.Site{ID: 3, Url: "http://test3.ru", IsRss: false}

	news1 := []repository.NewsItem{
		repository.NewsItem{Link: "http://test1.ru/news/1"},
		repository.NewsItem{Link: "http://test1.ru/news/2"},
	}
	news2 := []repository.NewsItem{
		repository.NewsItem{Link: "http://test2.ru/news/1"},
		repository.NewsItem{Link: "http://test2.ru/news/2"},
	}

	app.parser.(*mockedParser).
		On("Parse", site1).
		Return(news1, nil)
	app.parser.(*mockedParser).
		On("Parse", site2).
		Return(news2, nil)
	app.parser.(*mockedParser).
		On("Parse", site3).
		Return([]repository.NewsItem{}, errors.New("test parser error"))

	app.repository.(*mockedRepository).
		On("GetSites").
		Return([]repository.Site{site1, site2, site3}, nil)

	app.repository.(*mockedRepository).
		On("HasNewsItem", mock.MatchedBy(func(item repository.NewsItem) bool { return item.Link == news1[0].Link })).
		Return(false, nil)
	app.repository.(*mockedRepository).
		On("HasNewsItem", mock.MatchedBy(func(item repository.NewsItem) bool { return item.Link == news1[1].Link })).
		Return(true, nil)
	app.repository.(*mockedRepository).
		On("HasNewsItem", mock.MatchedBy(func(item repository.NewsItem) bool { return item.Link == news2[0].Link })).
		Return(false, nil)
	app.repository.(*mockedRepository).
		On("HasNewsItem", mock.MatchedBy(func(item repository.NewsItem) bool { return item.Link == news2[1].Link })).
		Return(false, errors.New("test repository error"))

	app.repository.(*mockedRepository).
		On("AddNewsItem", mock.MatchedBy(func(item *repository.NewsItem) bool { return item.Link == news1[0].Link })).
		Return(nil)
	app.repository.(*mockedRepository).
		On("AddNewsItem", mock.MatchedBy(func(item *repository.NewsItem) bool { return item.Link == news2[0].Link })).
		Return(errors.New("test repository error"))

	app.parsing()
	time.Sleep(time.Millisecond * 100)

	app.parser.(*mockedParser).AssertNumberOfCalls(t, "Parse", 3)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "HasNewsItem", 4)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "AddNewsItem", 2)

	time.Sleep(app.interval)

	app.parser.(*mockedParser).AssertNumberOfCalls(t, "Parse", 6)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "HasNewsItem", 8)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "AddNewsItem", 4)

	app.Stop()

	time.Sleep(app.interval)

	app.parser.(*mockedParser).AssertNumberOfCalls(t, "Parse", 6)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "HasNewsItem", 8)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "AddNewsItem", 4)
}
