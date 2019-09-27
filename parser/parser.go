package parser

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gaus57/news-agg/repository"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type HttpClient interface {
	Get(url string) (*http.Response, error)
}

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
	client HttpClient
}

func NewParser(client HttpClient) *parser {
	return &parser{
		client: client,
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
			item.Link = prepareLink(*response.Request.URL, link)
		}
		imageSelection := s.Find(imagePath)
		if image, ok := imageSelection.Attr("src"); ok {
			item.Image = prepareLink(*response.Request.URL, image)
		}
		news = append(news, item)
	})

	return
}

func (parser *parser) parseRss(response *http.Response) (news []repository.NewsItem, err error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

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

func prepareLink(siteUrl url.URL, link string) string {
	linkUrl, err := url.Parse(link)
	if err == nil {
		if linkUrl.Host == "" {
			linkUrl.Host = siteUrl.Host
			linkUrl.Scheme = siteUrl.Scheme
			if !strings.HasPrefix(link, "/") {
				if strings.HasSuffix(siteUrl.Path, "/") {
					linkUrl.Path = siteUrl.Path + link
				} else {
					linkUrl.Path = siteUrl.Path + "/" + link
				}
			}
			link = linkUrl.String()
		}
	}

	return link
}
