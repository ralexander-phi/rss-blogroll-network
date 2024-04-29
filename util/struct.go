package main

import (
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/rss"
)

type Config struct {
	Feeds            []FeedDetails
	FeedUrl          string   `yaml:"feed_url"`
	PostAgeLimitDays *int     `yaml:"post_age_limit_days"`
	MaxPostsPerFeed  *int     `yaml:"max_posts_per_feed"`
	MaxPosts         *int     `yaml:"max_posts"`
	BlockWords       []string `yaml:"block_words"`
	BlockDomains     []string `yaml:"block_domains"`
}

type FeedDetails struct {
	Link  string
	Text  string
	Title string
	Type  string
}

// https://gohugo.io/content-management/front-matter/
type PostFrontmatter struct {
	Date        string     `yaml:"date"`
	Params      PostParams `yaml:"params"`
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
}

type FeedFrontmatter struct {
	Date        string     `yaml:"date"`
	Params      FeedParams `yaml:"params"`
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
}

type PostParams struct {
	PrettyAge    string      `yaml:"pretty_age"`
	Post         gofeed.Item `yaml:"post"`
	FeedId       string      `yaml:"feed_id"`
	Feed         gofeed.Feed `yaml:"feed"`
	CommentsLink string      `yaml:"comments_link"`
}

type FeedParams struct {
	Id   string      `yaml:"id"`
	Feed gofeed.Feed `yaml:"feed"`
}

type MultiTypeItem struct {
	item *gofeed.Item
	rss  *rss.Item
}
