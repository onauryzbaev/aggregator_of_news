package parser

import (
	"errors"
	"fmt"
	"github.com/gaus57/news-agg/repository"
	"io/ioutil"
	"net/http"
)

type parser struct {
	client *http.Client
}

func NewParser() *parser {
	return &parser{
		client: &http.Client{},
	}
}

func (parser *parser) Parse(site repository.Site) (news []repository.NewsItem, err error) {
	response, err := parser.client.Get(site.Url)
	if err != nil {
		return
	}

	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("request failed with status code %d", response.StatusCode))

		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	if site.IsRss {
		news, err = parser.parseRss(body)
	} else {
		news, err = parser.parseHtml(body, site)
	}

	return
}
