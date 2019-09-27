package repository

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type Site struct {
	ID              int
	IsRss           bool   `gorm:"not null"`
	Url             string `gorm:"size:500;unique;not null"`
	NewsItemPath    string `gorm:"size:100"`
	TitlePath       string `gorm:"size:100"`
	DescriptionPath string `gorm:"size:100"`
	LinkPath        string `gorm:"size:100"`
	DatePath        string `gorm:"size:100"`
	ImagePath       string `gorm:"size:100"`
}

type NewsItem struct {
	ID          int
	SiteID      int    `gorm:"not null"`
	Site        Site   `gorm:"foreignkey:SiteRefer;association_autoupdate:false;association_autocreate:false"`
	Title       string `gorm:"size:250"`
	Description string
	Link        string `gorm:"size:500;unique;not null"`
	Date        string `gorm:"size:100"`
	Image       string `gorm:"size:500"`
}

type repository struct {
	conn *gorm.DB
}

func NewRepository(conn *gorm.DB) *repository {
	return &repository{conn}
}

func (rep *repository) Migrate() {
	rep.conn.AutoMigrate(&Site{}, &NewsItem{})
	rep.conn.Model(&NewsItem{}).AddForeignKey("site_id", "sites(id)", "CASCADE", "CASCADE")
}

func (rep *repository) GetSites() (sites []Site, err error) {
	err = rep.conn.Order("id desc").Find(&sites).Error

	return
}

func (rep *repository) AddSite(site *Site) error {
	return rep.conn.FirstOrCreate(site, Site{Url: site.Url}).Error
}

func (rep *repository) DeleteSite(id int) error {
	return rep.conn.Delete(&Site{ID: id}).Error
}

func (rep *repository) GetNews(offset int, limit int, search string) (news []NewsItem, err error) {
	q := rep.conn.Order("id desc").Offset(offset).Limit(limit)
	if search != "" {
		q = q.Where("title ILIKE ?", fmt.Sprintf("%%%s%%", search))
	}
	err = q.Find(&news).Error

	return
}

func (rep *repository) HasNewsItem(item NewsItem) (bool, error) {
	q := rep.conn.Where("link = ?", item.Link).First(&NewsItem{})
	if q.RecordNotFound() {
		return false, nil
	} else if q.Error != nil {
		return false, q.Error
	}

	return true, nil
}

func (rep *repository) AddNewsItem(item *NewsItem) error {
	return rep.conn.Create(item).Error
}
