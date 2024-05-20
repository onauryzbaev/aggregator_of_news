package main

import (
	"errors"
	"github.com/onauryzbaev/go_news_final_/tree/master/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
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

	time.Sleep(app.interval)

	app.parser.(*mockedParser).AssertNumberOfCalls(t, "Parse", 9)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "HasNewsItem", 12)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "AddNewsItem", 6)

	app.Stop()

	time.Sleep(app.interval)

	app.parser.(*mockedParser).AssertNumberOfCalls(t, "Parse", 9)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "HasNewsItem", 12)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "AddNewsItem", 6)
}

func TestMainHandler(t *testing.T) {
	app := getApplication()
	app.prepareTemplates()

	app.repository.(*mockedRepository).
		On("GetNews", 0, 10, "").
		Return(
			[]repository.NewsItem{
				repository.NewsItem{
					Title:       "Заголовок 1",
					Link:        "http://test1.ru/news/1",
					Description: "описание 1",
					Date:        "2019-09-27 04:00",
					Image:       "http://test1.ru/news/1.jpeg",
				},
				repository.NewsItem{
					Title:       "Заголовок 2",
					Link:        "http://test1.ru/news/2",
					Description: "описание 2",
					Date:        "2019-09-27 05:00",
					Image:       "http://test1.ru/news/2.jpeg",
				},
			},
			nil,
		)
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.mainHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Заголовок 1")
	assert.Contains(t, rr.Body.String(), "Заголовок 2")

	app.repository.(*mockedRepository).
		On("GetNews", 10, 10, "поиск").
		Return([]repository.NewsItem{}, errors.New("test repository error"))
	req, _ = http.NewRequest("GET", "/?q=поиск&page=2", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	req, _ = http.NewRequest("GET", "/not-found", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSitesHandler(t *testing.T) {
	app := getApplication()
	app.prepareTemplates()

	app.repository.(*mockedRepository).
		On("GetSites").
		Return(
			[]repository.Site{
				repository.Site{
					ID:  1,
					Url: "http://test1.ru/news/",
				},
				repository.Site{
					ID:  1,
					Url: "http://test2.ru/news/",
				},
			},
			nil,
		)
	req, _ := http.NewRequest("GET", "/sites", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.sitesHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "http://test1.ru/news/")
	assert.Contains(t, rr.Body.String(), "http://test2.ru/news/")

	app = getApplication()
	app.prepareTemplates()
	app.repository.(*mockedRepository).
		On("GetSites").
		Return([]repository.Site{}, errors.New("test repository error"))
	req, _ = http.NewRequest("GET", "/sites", nil)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(app.sitesHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSiteDeleteHandler(t *testing.T) {
	app := getApplication()
	app.prepareTemplates()

	app.repository.(*mockedRepository).
		On("DeleteSite", 1).
		Return(nil)
	reader := strings.NewReader("id=1")
	req, _ := http.NewRequest("POST", "/sites/delete", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.siteDeleteHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "DeleteSite", 1)

	app.repository.(*mockedRepository).
		On("DeleteSite", 2).
		Return(errors.New("test repository error"))
	reader = strings.NewReader("id=2")
	req, _ = http.NewRequest("POST", "/sites/delete", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "DeleteSite", 2)

	reader = strings.NewReader("id=invalid")
	req, _ = http.NewRequest("POST", "/sites/delete", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "DeleteSite", 2)

	req, _ = http.NewRequest("GET", "/sites/delete", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestSiteAddHandler(t *testing.T) {
	app := getApplication()
	app.prepareTemplates()

	req, _ := http.NewRequest("GET", "/sites/add", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.siteAddHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Добавить сайт")

	app.repository.(*mockedRepository).
		On("AddSite", &repository.Site{Url: "http://test1.ru", IsRss: true}).
		Return(nil)
	reader := strings.NewReader("url=http://test1.ru;is_rss=1")
	req, _ = http.NewRequest("POST", "/sites/add", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "AddSite", 1)

	app.repository.(*mockedRepository).
		On("AddSite", &repository.Site{
			Url:             "http://test2.ru",
			IsRss:           false,
			NewsItemPath:    "article",
			TitlePath:       "h3",
			DescriptionPath: ".desc",
			LinkPath:        "a",
			DatePath:        "i",
			ImagePath:       "img",
		}).
		Return(errors.New("test repository error"))
	reader = strings.NewReader("url=http://test2.ru;is_rss=0;news_item_path=article;title_path=h3;description_path=.desc;link_path=a;date_path=i;image_path=img")
	req, _ = http.NewRequest("POST", "/sites/add", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	app.repository.(*mockedRepository).AssertNumberOfCalls(t, "AddSite", 2)
}
