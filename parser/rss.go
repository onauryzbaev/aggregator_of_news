package parser

import (
	"encoding/xml"
	"fmt"
	"github.com/gaus57/news-agg/repository"
)

func (parser *parser) parseRss(content []byte) (news []repository.NewsItem, err error) {
	fmt.Println("content rss", string(content))
	v := new(interface{})
	err = xml.Unmarshal(content, v)
	if err != nil {
		return
	}

	return
}
