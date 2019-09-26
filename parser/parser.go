package parser

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gaus57/news-agg/repository"
	"io/ioutil"
	"net/http"
)

type rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	XMLName xml.Name `xml:"channel"`
	Items   []item   `xml:"item"`
}

type item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
	Date        string   `xml:"pubDate"`
	Image       string   `xml:"image"`
}

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

	if site.IsRss {
		news, err = parser.parseRss(response)
	} else {
		news, err = parser.parseHtml(
			response,
			site.NewsItemPath,
			site.TitlePath,
			site.DescriptionPath,
			site.LinkPath,
			site.DatePath,
			site.ImagePath,
		)
	}

	return
}

func (parser *parser) parseHtml(
	response *http.Response,
	itemPath,
	titlePath,
	descriptionPath,
	linkPath,
	datePath,
	imagePath string,
) (news []repository.NewsItem, err error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return
	}

	doc.Find(itemPath).Each(func(i int, s *goquery.Selection) {
		item := repository.NewsItem{
			Title:       s.Find(titlePath).Text(),
			Description: s.Find(descriptionPath).Text(),
			Date:        s.Find(datePath).Text(),
		}
		linkSelection := s.Find(linkPath)
		if link, ok := linkSelection.Attr("href"); ok {
			item.Link = link
		}
		imageSelection := s.Find(imagePath)
		if image, ok := imageSelection.Attr("src"); ok {
			item.Image = image
		}
		news = append(news, item)
	})

	return
}

func (parser *parser) parseRss(response *http.Response) (news []repository.NewsItem, err error) {
	body, err := ioutil.ReadAll(response.Body)
	rss := &rss{}
	err = xml.Unmarshal(body, rss)
	if err != nil {
		return
	}
	for _, rssItem := range rss.Channel.Items {
		news = append(news, repository.NewsItem{
			Title:       rssItem.Title,
			Description: rssItem.Description,
			Link:        rssItem.Link,
			Date:        rssItem.Date,
			Image:       rssItem.Image,
		})
	}

	return
}
