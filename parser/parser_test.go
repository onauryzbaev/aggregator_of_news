package parser

import (
	"errors"
	"github.com/gaus57/news-agg/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewParser(t *testing.T) {
	mockedClient := &mockedHttpClient{}
	parser := NewParser(mockedClient)
	assert.NotNil(t, parser)
}

func TestParse(t *testing.T) {
	t.Run("Http request error", func(t *testing.T) {
		mockedClient := &mockedHttpClient{}
		mockedClient.
			On("Get", "http://error.ru").
			Return(&http.Response{}, errors.New("test http error"))

		parser := NewParser(mockedClient)
		_, err := parser.Parse(repository.Site{Url: "http://error.ru"})
		assert.Error(t, err)
		assert.Equal(t, "test http error", err.Error())
	})

	t.Run("Http response with wrong status", func(t *testing.T) {
		mockedClient := &mockedHttpClient{}
		mockedClient.
			On("Get", "http://error-500.ru").
			Return(&http.Response{StatusCode: http.StatusInternalServerError}, nil)

		parser := NewParser(mockedClient)
		_, err := parser.Parse(repository.Site{Url: "http://error-500.ru"})
		assert.Error(t, err)
		assert.Equal(t, "request failed with status code 500", err.Error())
	})

	t.Run("Parse rss invalid response", func(t *testing.T) {
		mockedClient := &mockedHttpClient{}
		response := &httptest.ResponseRecorder{Code: http.StatusOK}
		mockedClient.
			On("Get", "http://invalid-xml-rss.ru").
			Return(response.Result(), nil)

		parser := NewParser(mockedClient)
		_, err := parser.Parse(repository.Site{Url: "http://invalid-xml-rss.ru", IsRss: true})
		assert.Error(t, err)
		assert.Equal(t, "EOF", err.Error())
	})

	t.Run("Parse html invalid response", func(t *testing.T) {
		mockedClient := &mockedHttpClient{}
		response := &httptest.ResponseRecorder{Code: http.StatusOK}
		mockedClient.
			On("Get", "http://invalid.ru").
			Return(response.Result(), nil)

		parser := NewParser(mockedClient)
		news, err := parser.Parse(repository.Site{Url: "http://invalid.ru", IsRss: false})
		assert.NoError(t, err)
		assert.Empty(t, news)
	})

	t.Run("Parse rss success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			_, _ = rw.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
				<rss version="2.0">
					<channel>
						<item>
							<title>Заголовок 1</title>
							<link>https://news.ru/1</link>
							<description>Описание 1</description>
							<pubDate>Wed, 25 Sep 2019 18:10:34 +0300</pubDate>
						</item>
						<item>
							<title>Заголовок 2</title>
							<link>https://news.ru/2</link>
							<description>Описание 2</description>
							<image>https://news.ru/2.jpeg</image>
						</item>
					</channel>
				</rss>
			`))
		}))
		defer server.Close()

		parser := NewParser(server.Client())
		news, err := parser.Parse(repository.Site{Url: server.URL, IsRss: true})
		assert.NoError(t, err)
		assert.Len(t, news, 2)
		assert.Equal(t, news[0], repository.NewsItem{
			Title:       "Заголовок 1",
			Description: "Описание 1",
			Date:        "Wed, 25 Sep 2019 18:10:34 +0300",
			Link:        "https://news.ru/1",
		})
		assert.Equal(t, news[1], repository.NewsItem{
			Title:       "Заголовок 2",
			Description: "Описание 2",
			Link:        "https://news.ru/2",
			Image:       "https://news.ru/2.jpeg",
		})
	})

	t.Run("Parse html success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			_, _ = rw.Write([]byte(`<!DOCTYPE html>
				<html>
					<body>
						<article class="art">
							<a href="https://news.ru/1"><h3>Заголовок 1</h3></a>
							<div class="art-desc">Описание 1</div>
							<i>Wed, 25 Sep 2019 18:10:34 +0300</i>
						</article>
						<article class="art">
							<a href="https://news.ru/2"><h3>Заголовок 2</h3></a>
							<div class="art-desc">Описание 2</div>
							<img src="https://news.ru/2.jpeg"/>
						</article>
					</body>
				</html>
			`))
		}))
		defer server.Close()

		parser := NewParser(server.Client())
		news, err := parser.Parse(repository.Site{
			Url:             server.URL,
			IsRss:           false,
			NewsItemPath:    "article.art",
			TitlePath:       "h3",
			DescriptionPath: ".art-desc",
			LinkPath:        "a",
			DatePath:        "",
			ImagePath:       "img",
		})
		assert.NoError(t, err)
		assert.Len(t, news, 2)
		assert.Equal(t, news[0], repository.NewsItem{
			Title:       "Заголовок 1",
			Description: "Описание 1",
			Date:        "",
			Link:        "https://news.ru/1",
		})
		assert.Equal(t, news[1], repository.NewsItem{
			Title:       "Заголовок 2",
			Description: "Описание 2",
			Link:        "https://news.ru/2",
			Date:        "",
			Image:       "https://news.ru/2.jpeg",
		})
	})
}

func TestPrepareLink(t *testing.T) {
	u := &url.URL{}

	t.Run("Prepare absolute link", func(t *testing.T) {
		siteUrl, _ := u.Parse("http://site.ru")
		link := prepareLink(*siteUrl, "http://site.ru/news/1")
		assert.Equal(t, "http://site.ru/news/1", link)
	})

	t.Run("Prepare relative link", func(t *testing.T) {
		siteUrl, _ := u.Parse("http://site.ru")
		link := prepareLink(*siteUrl, "/news/1")
		assert.Equal(t, "http://site.ru/news/1", link)

		siteUrl, _ = u.Parse("http://site.ru/news")
		link = prepareLink(*siteUrl, "1")
		assert.Equal(t, "http://site.ru/news/1", link)

		siteUrl, _ = u.Parse("http://site.ru/news")
		link = prepareLink(*siteUrl, "/news/1")
		assert.Equal(t, "http://site.ru/news/1", link)
	})
}

type mockedHttpClient struct {
	mock.Mock
}

func (client *mockedHttpClient) Get(url string) (*http.Response, error) {
	calArgs := client.MethodCalled("Get", url)

	return calArgs.Get(0).(*http.Response), calArgs.Error(1)
}
