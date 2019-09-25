package repository

type Site struct {
	ID                 int
	IsRss              bool
	Url                string
	Name               string
	ItemsContainerPath string
	TitlePath          string
	DescriptionPath    string
	LinkPath           string
	DatePath           string
	ImagePath          string
}

type NewsItem struct {
	ID          int
	SiteID      int
	Title       string
	Description string
	Link        string
	Date        string
	Image       string
}
