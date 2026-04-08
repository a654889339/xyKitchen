package models

import "time"

type ServiceCategory struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Bg        *string   `gorm:"size:50" json:"bg"`
	BgOpacity *float64  `gorm:"column:bgOpacity;type:decimal(5,2)" json:"bgOpacity"`
	Key       *string   `gorm:"size:50" json:"key"`
	NameEn    string    `gorm:"size:100" json:"nameEn"`
	SortOrder int       `gorm:"column:sortOrder" json:"sortOrder"`
	Status    string    `gorm:"type:enum('active','inactive');default:active" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Services  []Service `gorm:"foreignKey:CategoryID" json:"services,omitempty"`
}

func (ServiceCategory) TableName() string { return "service_categories" }

type Service struct {
	ID             int      `gorm:"primaryKey" json:"id"`
	Title          string   `gorm:"size:200;not null" json:"title"`
	Description    string   `gorm:"type:text" json:"description"`
	Icon           string   `gorm:"size:500" json:"icon"`
	IconURL        string   `gorm:"column:iconUrl;size:500" json:"iconUrl"`
	Cover          string   `gorm:"size:500" json:"cover"`
	Category       *string  `gorm:"size:50" json:"category"`
	CategoryID     *int     `gorm:"column:categoryId" json:"categoryId"`
	Price          float64  `gorm:"type:decimal(10,2)" json:"price"`
	OriginPrice    *float64 `gorm:"column:originPrice;type:decimal(10,2)" json:"originPrice"`
	Bg             string   `gorm:"size:50" json:"bg"`
	BgOpacity      *float64 `gorm:"column:bgOpacity;type:decimal(5,2)" json:"bgOpacity"`
	TitleEn        string   `gorm:"column:titleEn;size:200" json:"titleEn"`
	DescriptionEn  string   `gorm:"column:descriptionEn;type:text" json:"descriptionEn"`
	PriceEn        *float64 `gorm:"column:priceEn;type:decimal(10,2)" json:"priceEn"`
	OriginPriceEn  *float64 `gorm:"column:originPriceEn;type:decimal(10,2)" json:"originPriceEn"`
	CurrencyEn     string   `gorm:"column:currencyEn;size:10" json:"currencyEn"`
	Status         string   `gorm:"type:enum('active','inactive');default:active" json:"status"`
	SortOrder      int      `gorm:"column:sortOrder" json:"sortOrder"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	ServiceCategory *ServiceCategory `gorm:"foreignKey:CategoryID" json:"serviceCategory,omitempty"`
}

func (Service) TableName() string { return "services" }

type ProductCategory struct {
	ID              int       `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	ThumbnailURL    *string   `gorm:"column:thumbnail_url;size:1024" json:"thumbnailUrl"`
	NameEn          string    `gorm:"size:100" json:"nameEn"`
	ThumbnailURLEn  *string   `gorm:"column:thumbnailUrlEn;size:1024" json:"thumbnailUrlEn"`
	SortOrder       int       `gorm:"column:sortOrder" json:"sortOrder"`
	Status          string    `gorm:"type:enum('active','inactive');default:active" json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (ProductCategory) TableName() string { return "product_categories" }

type HomeConfig struct {
	ID               int       `gorm:"primaryKey" json:"id"`
	Section          string    `gorm:"size:50;not null" json:"section"`
	Title            string    `gorm:"size:100" json:"title"`
	Desc             string    `gorm:"size:200" json:"desc"`
	Icon             string    `gorm:"size:100" json:"icon"`
	ImageURL         string    `gorm:"column:imageUrl;size:500" json:"imageUrl"`
	ImageURLThumb    string    `gorm:"column:imageUrlThumb;size:500" json:"imageUrlThumb"`
	Color            string    `gorm:"size:200" json:"color"`
	Path             string    `gorm:"size:200" json:"path"`
	Price            string    `gorm:"size:20" json:"price"`
	TitleEn          string    `gorm:"size:100" json:"titleEn"`
	DescEn           string    `gorm:"size:200" json:"descEn"`
	IconEn           string    `gorm:"size:100" json:"iconEn"`
	ImageURLEn       string    `gorm:"column:imageUrlEn;size:500" json:"imageUrlEn"`
	ImageURLThumbEn  string    `gorm:"column:imageUrlThumbEn;size:500" json:"imageUrlThumbEn"`
	SortOrder        int       `gorm:"column:sortOrder" json:"sortOrder"`
	Status           string    `gorm:"type:enum('active','inactive');default:active" json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (HomeConfig) TableName() string { return "home_configs" }

type OutletServiceCategory struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Bg        *string   `gorm:"size:50" json:"bg"`
	BgOpacity *float64  `gorm:"column:bgOpacity;type:decimal(5,2)" json:"bgOpacity"`
	Key       *string   `gorm:"size:50" json:"key"`
	SortOrder int       `gorm:"column:sortOrder" json:"sortOrder"`
	Status    string    `gorm:"type:enum('active','inactive');default:active" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Services  []OutletService `gorm:"foreignKey:CategoryID" json:"services,omitempty"`
}

func (OutletServiceCategory) TableName() string { return "outlet_service_categories" }

type OutletService struct {
	ID          int      `gorm:"primaryKey" json:"id"`
	Title       string   `gorm:"size:200;not null" json:"title"`
	Description string   `gorm:"type:text" json:"description"`
	Icon        string   `gorm:"size:500" json:"icon"`
	IconURL     string   `gorm:"column:iconUrl;size:500" json:"iconUrl"`
	Cover       string   `gorm:"size:500" json:"cover"`
	Category    *string  `gorm:"size:50" json:"category"`
	CategoryID  *int     `gorm:"column:categoryId" json:"categoryId"`
	Price       float64  `gorm:"type:decimal(10,2)" json:"price"`
	OriginPrice *float64 `gorm:"column:originPrice;type:decimal(10,2)" json:"originPrice"`
	Bg          string   `gorm:"size:50" json:"bg"`
	BgOpacity   *float64 `gorm:"column:bgOpacity;type:decimal(5,2)" json:"bgOpacity"`
	Status      string   `gorm:"type:enum('active','inactive');default:active" json:"status"`
	SortOrder   int      `gorm:"column:sortOrder" json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ServiceCategory *OutletServiceCategory `gorm:"foreignKey:CategoryID" json:"serviceCategory,omitempty"`
}

func (OutletService) TableName() string { return "outlet_services" }

type OutletHomeConfig struct {
	ID            int       `gorm:"primaryKey" json:"id"`
	Section       string    `gorm:"size:50;not null" json:"section"`
	Title         string    `gorm:"size:100" json:"title"`
	Desc          string    `gorm:"size:200" json:"desc"`
	Icon          string    `gorm:"size:100" json:"icon"`
	ImageURL      string    `gorm:"column:imageUrl;size:500" json:"imageUrl"`
	ImageURLThumb string    `gorm:"column:imageUrlThumb;size:500" json:"imageUrlThumb"`
	Color         string    `gorm:"size:200" json:"color"`
	Path          string    `gorm:"size:200" json:"path"`
	Price         string    `gorm:"size:20" json:"price"`
	SortOrder     int       `gorm:"column:sortOrder" json:"sortOrder"`
	Status        string    `gorm:"type:enum('active','inactive');default:active" json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (OutletHomeConfig) TableName() string { return "outlet_home_configs" }
