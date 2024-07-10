package main

type Feed struct {
	ID          uint   `gorm:"primaryKey"`
	Date        string // TODO: use time.Time
	Description string
	Title       string
	FeedLink    string
	FeedId      string
	FeedType    string
	IsPodcast   bool
	IsNoarchive bool
}

type Post struct {
	ID          uint   `gorm:"primaryKey"`
	Date        string // TODO: use time.Time
	Description string
	Title       string
	FeedId      string
	PostLink    string
	Guid        string
}

type PostsByCategory struct {
	ID       uint `gorm:"primaryKey"`
	Category string
	Link     string
}

type PostsByLanguage struct {
	ID       uint `gorm:"primaryKey"`
	Language string
	Link     string
}

type FeedsByCategory struct {
	ID       uint `gorm:"primaryKey"`
	Category string
	Link     string
}

type FeedsByLanguage struct {
	ID       uint `gorm:"primaryKey"`
	Language string
	Link     string
}

type Link struct {
	ID              uint `gorm:"primaryKey"`
	SourceType      int
	SourceUrl       string
	DestinationType int
	DestinationUrl  string
	LinkType        string
}
