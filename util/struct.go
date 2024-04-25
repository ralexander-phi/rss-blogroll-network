package main

import (
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/rss"
)

type FeedConfig struct {
	URL               string   `yaml:"url"`
	IgnoreDescription *bool    `yaml:"ignore_description"`
	IgnoreContent     *bool    `yaml:"ignore_content"`
	BlockWords        []string `yaml:"block_words"`
	BlockDomains      []string `yaml:"block_domains"`
}

type Config struct {
	Feeds            []FeedConfig `yaml:"feeds"`
	PostAgeLimitDays *int         `yaml:"post_age_limit_days"`
	MaxPostsPerFeed  *int         `yaml:"max_posts_per_feed"`
	MaxPosts         *int         `yaml:"max_posts"`
	BlockWords       []string     `yaml:"block_words"`
	BlockDomains     []string     `yaml:"block_domains"`
}

// https://gohugo.io/content-management/front-matter/
type Frontmatter struct {
	Date        string     `yaml:"date"`
	Params      PostParams `yaml:"params"`
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
}

type PostParams struct {
	PrettyAge    string      `yaml:"pretty_age"`
	Post         gofeed.Item `yaml:"post"`
	Feed         gofeed.Feed `yaml:"feed"`
	CommentsLink string      `yaml:"comments_link"`
}

type MultiTypeItem struct {
	feed *gofeed.Feed
	item *gofeed.Item
	rss  *rss.Item
}
