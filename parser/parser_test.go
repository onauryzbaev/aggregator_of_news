package parser

import (
	"errors"
	"github.com/gaus57/news-agg/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
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
		_, _ = response.WriteString(`server error`)
		mockedClient.
			On("Get", "http://invalid-xml-rss.ru").
			Return(response.Result(), nil)

		parser := NewParser(mockedClient)
		_, err := parser.Parse(repository.Site{Url: "http://invalid-xml-rss.ru", IsRss: true})
		assert.Error(t, err)
		assert.Equal(t, "EOF", err.Error())
	})
}

type mockedHttpClient struct {
	mock.Mock
}

func (client *mockedHttpClient) Get(url string) (*http.Response, error) {
	calArgs := client.MethodCalled("Get", url)

	return calArgs.Get(0).(*http.Response), calArgs.Error(1)
}
