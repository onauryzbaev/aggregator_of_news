package main

import (
	"github.com/gaus57/news-agg/repository"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

type Repository interface {
	Migrate()
	GetSites() ([]repository.Site, error)
	AddSite(site *repository.Site) error
	DeleteSite(id int) error
	GetNews(offset int, limit int, search string) ([]repository.NewsItem, error)
	HasNewsItem(item repository.NewsItem) (bool, error)
	AddNewsItem(item *repository.NewsItem) error
}

type Parser interface {
	Parse(site repository.Site) ([]repository.NewsItem, error)
}

type application struct {
	log        *log.Logger
	repository Repository
	parser     Parser
	stop       chan bool
	templates  *template.Template
}

func NewApplication(repository Repository, parser Parser, logger *log.Logger) *application {
	return &application{
		log:        logger,
		repository: repository,
		parser:     parser,
		stop:       make(chan bool),
	}
}

func (app *application) Serve() {
	app.repository.Migrate()
	app.parsing()
	app.serveHttp()
}

func (app *application) Stop() {
	close(app.stop)
}

func (app *application) parsing() {
	go func() {
		app.parseSites()
		ticker := time.NewTicker(time.Minute * 10)
		select {
		case <-app.stop:
			return
		case <-ticker.C:
			app.parseSites()
		}
	}()
}

func (app *application) parseSites() {
	sites, err := app.repository.GetSites()
	if err != nil {
		app.log.Printf("Failed get sites from repository: %v", err)

		return
	}

	for _, site := range sites {
		news, err := app.parser.Parse(site)
		if err != nil {
			app.log.Printf("Failed parse site %s: %v", site.Url, err)
			continue
		}

		insert := 0
		for _, item := range news {
			item.SiteID = site.ID
			exist, err := app.repository.HasNewsItem(item)
			if err != nil {
				app.log.Printf("Failed to check exist news in repository: %v", err)
				continue
			}
			if exist {
				continue
			}
			err = app.repository.AddNewsItem(&item)
			if err != nil {
				app.log.Printf("Failed add news to repository: %v", err)
				continue
			}
			insert++
		}

		app.log.Printf("Complete parse site %s - add %d news", site.Url, insert)
	}
}

func (app *application) prepareTemplates() {
	var allFiles []string
	files, err := ioutil.ReadDir("./tmpl")
	if err != nil {
		app.log.Printf("Fail read tmpl dir: %v", err)

		return
	}

	for _, file := range files {
		allFiles = append(allFiles, "./tmpl/"+file.Name())
	}
	app.templates, err = template.ParseFiles(allFiles...)
	if err != nil {
		app.log.Printf("Fail parse templates: %v", err)
	}
}

func (app *application) serveHttp() {
	app.prepareTemplates()

	http.HandleFunc("/", app.mainHandler)
	http.HandleFunc("/sites", app.sitesHandler)
	http.HandleFunc("/sites/add", app.siteAddHandler)
	http.HandleFunc("/sites/delete", app.siteDeleteHandler)

	app.log.Fatal(http.ListenAndServe(":8080", nil))
}

func (app *application) mainHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte("Страница не найдена"))

		return
	}

	search := req.URL.Query().Get("q")
	page, err := strconv.Atoi(req.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}
	perpage := 10
	news, err := app.repository.GetNews((page-1)*perpage, perpage, search)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		app.log.Printf("Fail get news from repository: %v", err)

		return
	}

	tmpl := app.templates.Lookup("main.tmpl")
	err = tmpl.Execute(
		res,
		struct {
			NewsItems []repository.NewsItem
			Search    string
			Page      int
			PrevPage  int
			NextPage  int
		}{
			news, search, page, page - 1, page + 1,
		},
	)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		app.log.Printf("Fail execute template: %v", err)

		return
	}
	res.WriteHeader(http.StatusOK)
}

func (app *application) sitesHandler(res http.ResponseWriter, req *http.Request) {
	sites, err := app.repository.GetSites()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		app.log.Printf("Fail get sites from repository: %v", err)

		return
	}

	tmpl := app.templates.Lookup("sites.tmpl")
	err = tmpl.Execute(
		res,
		struct{ Sites []repository.Site }{
			sites,
		},
	)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		app.log.Printf("Fail execute template: %v", err)

		return
	}
	res.WriteHeader(http.StatusOK)
}

func (app *application) siteDeleteHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	id, err := strconv.Atoi(req.FormValue("id"))
	if err != nil {
		http.Redirect(res, req, "/sites", http.StatusTemporaryRedirect)

		return
	}

	err = app.repository.DeleteSite(id)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		app.log.Printf("Fail delete site from repository: %v", err)

		return
	}

	http.Redirect(res, req, "/sites", http.StatusTemporaryRedirect)
}

func (app *application) siteAddHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		app.siteCreateHandler(res, req)

		return
	}

	tmpl := app.templates.Lookup("site_add.tmpl")
	err := tmpl.Execute(res, nil)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		app.log.Printf("Fail execute template: %v", err)

		return
	}

	res.WriteHeader(http.StatusOK)
}

func (app *application) siteCreateHandler(res http.ResponseWriter, req *http.Request) {
	site := &repository.Site{}
	site.IsRss, _ = strconv.ParseBool(req.FormValue("is_rss"))
	site.Url = req.FormValue("url")
	site.NewsItemPath = req.FormValue("news_item_path")
	site.TitlePath = req.FormValue("title_path")
	site.LinkPath = req.FormValue("link_path")
	site.DescriptionPath = req.FormValue("description_path")
	site.DatePath = req.FormValue("date_path")
	site.ImagePath = req.FormValue("image_path")

	err := app.repository.AddSite(site)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		app.log.Printf("Fail insert site to repository: %v", err)

		return
	}

	http.Redirect(res, req, "/sites", http.StatusTemporaryRedirect)
}
