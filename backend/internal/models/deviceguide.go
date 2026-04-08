package models

import (
	"time"

	"gorm.io/datatypes"
)

type DeviceGuide struct {
	ID                 int            `gorm:"primaryKey" json:"id"`
	Name               string         `gorm:"size:100;not null" json:"name"`
	Slug               *string        `gorm:"size:100;uniqueIndex:slug" json:"slug"`
	Subtitle           string         `gorm:"size:200" json:"subtitle"`
	Icon               string         `gorm:"size:100" json:"icon"`
	IconURL            string         `gorm:"column:iconUrl;size:500" json:"iconUrl"`
	IconURLThumb       string         `gorm:"column:iconUrlThumb;size:500" json:"iconUrlThumb"`
	Emoji              string         `gorm:"size:20" json:"emoji"`
	Gradient           string         `gorm:"size:300" json:"gradient"`
	Badge              string         `gorm:"size:20" json:"badge"`
	Tags               datatypes.JSON `gorm:"type:text" json:"tags"`
	Sections           datatypes.JSON `gorm:"type:longtext" json:"sections"`
	CoverImage         string         `gorm:"column:coverImage;size:500" json:"coverImage"`
	CoverImageThumb    string         `gorm:"column:coverImageThumb;size:500" json:"coverImageThumb"`
	ShowcaseVideo      string         `gorm:"column:showcaseVideo;size:500" json:"showcaseVideo"`
	Description        string         `gorm:"type:text" json:"description"`
	MediaItems         datatypes.JSON `gorm:"type:longtext;column:mediaItems" json:"mediaItems"`
	HelpItems          datatypes.JSON `gorm:"type:longtext;column:helpItems" json:"helpItems"`
	ManualPdfURL       string         `gorm:"column:manualPdfUrl;size:500" json:"manualPdfUrl"`
	CategoryID         *int           `gorm:"column:categoryId" json:"categoryId"`
	SortOrder          int            `gorm:"column:sortOrder" json:"sortOrder"`
	QrcodeURL          string         `gorm:"column:qrcodeUrl;size:500" json:"qrcodeUrl"`
	NameEn             string         `gorm:"size:100" json:"nameEn"`
	SubtitleEn         string         `gorm:"size:200" json:"subtitleEn"`
	BadgeEn            string         `gorm:"size:20" json:"badgeEn"`
	DescriptionEn      string         `gorm:"type:text" json:"descriptionEn"`
	IconURLEn          string         `gorm:"column:iconUrlEn;size:500" json:"iconUrlEn"`
	IconURLThumbEn     string         `gorm:"column:iconUrlThumbEn;size:500" json:"iconUrlThumbEn"`
	CoverImageEn       string         `gorm:"column:coverImageEn;size:500" json:"coverImageEn"`
	CoverImageThumbEn  string         `gorm:"column:coverImageThumbEn;size:500" json:"coverImageThumbEn"`
	EmojiEn            string         `gorm:"size:20" json:"emojiEn"`
	GradientEn         string         `gorm:"size:300" json:"gradientEn"`
	Status             string         `gorm:"type:enum('active','inactive');default:active" json:"status"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	Category           *ProductCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (DeviceGuide) TableName() string { return "device_guides" }
