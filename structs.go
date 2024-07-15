package main

import (
	"github.com/go-yaml/yaml"
	"os"
)

type BlogrollInfo struct {
	Title       string             `yaml:"title"`
	Date        string             `yaml:"date"`
	Description string             `yaml:"description"`
	Params      BlogrollInfoParams `yaml:"params"`
}

type BlogrollInfoParams struct {
	Link       string                   `yaml:"link"`
	BlogrollId string                   `yaml:"blogroll_id"`
	Recommends []BlogrollRecommendation `yaml:"recommends"`
}

type BlogrollRecommendation struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Id          string `yaml:"id"`
}

func NewBlogrollInfo(row Blogroll) *BlogrollInfo {
	params := BlogrollInfoParams{
		Link:       row.Link,
		BlogrollId: row.BlogrollId,
	}
	return &BlogrollInfo{
		Title:       row.Title,
		Date:        row.Date,
		Description: row.Description,
		Params:      params,
	}
}

func (f *BlogrollInfo) Save() {
	output, err := yaml.Marshal(f)
	ohno(err)

	// Markdown uses `---` for YAML frontmatter
	sep := []byte("---\n")
	output = append(sep, output...)
	output = append(output, sep...)

	path := "content/blogrolls/br-" + f.Params.BlogrollId + ".md"
	err = os.WriteFile(path, output, os.FileMode(int(0660)))
	ohno(err)
}

type FeedInfo struct {
	Title       string         `yaml:"title"`
	Date        string         `yaml:"date"`
	Description string         `yaml:"description"`
	Params      FeedInfoParams `yaml:"params"`
}

type BlogrollBrief struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Id          string `yaml:"id"`
}

type FeedInfoParams struct {
	FeedLink           string          `yaml:"feedlink"`
	FeedType           string          `yaml:"feedtype"`
	FeedID             string          `yaml:"feedid"`
	Websites           map[string]bool `yaml:"websites"`
	Blogrolls          []string        `yaml:"blogrolls"`
	InBlogroll         []BlogrollBrief `yaml:"in_blogrolls"`
	Recommended        []string        `yaml:"recommended"`
	Recommender        []string        `yaml:"recommender"`
	Categories         []string        `yaml:"categories"`
	RelMe              map[string]bool `yaml:"relme"`
	LastPostTitle      string          `yaml:"last_post_title"`
	LastPostDesc       string          `yaml:"last_post_description"`
	LastPostDate       string          `yaml:"last_post_date"`
	LastPostLink       string          `yaml:"last_post_link"`
	LastPostCategories []string        `yaml:"last_post_categories"`
	LastPostLanguage   string          `yaml:"last_post_language"`
	LastPostGuid       string          `yaml:"last_post_guid"`
	ScoreCriteria      map[string]int  `yaml:"score_criteria"`
	Score              int             `yaml:"score"`
	IsPodcast          bool            `yaml:"ispodcast"`
	IsNoarchive        bool            `yaml:"isnoarchive"`
	InNetwork          bool            `yaml:"innetwork"`
	Language           string          `yaml:"language"`
}

func NewFeedInfo(row Feed) *FeedInfo {
	p := FeedInfoParams{
		FeedLink:      row.FeedLink,
		FeedID:        row.FeedId,
		FeedType:      row.FeedType,
		Websites:      map[string]bool{},
		RelMe:         map[string]bool{},
		ScoreCriteria: map[string]int{},
		IsPodcast:     row.IsPodcast,
		IsNoarchive:   row.IsNoarchive,
	}
	f := FeedInfo{
		Title:       row.Title,
		Date:        row.Date,
		Description: row.Description,
		Params:      p,
	}
	return &f
}

type RejectInfo struct {
	Title  string           `yaml:"title"`
	Params RejectInfoParams `yaml:"params"`
}

type RejectInfoParams struct {
	FeedLink string `yaml:"feedlink"`
	FeedID   string `yaml:"feedid"`
}

func (f *FeedInfo) Reject() {
	r := RejectInfo{
		Title: f.Title,
		Params: RejectInfoParams{
			FeedLink: f.Params.FeedLink,
			FeedID:   f.Params.FeedID,
		},
	}

	output, err := yaml.Marshal(r)
	ohno(err)

	// Markdown uses `---` for YAML frontmatter
	sep := []byte("---\n")
	output = append(sep, output...)
	output = append(output, sep...)

	path := "content/excluded/feed-" + r.Params.FeedID + ".md"
	err = os.WriteFile(path, output, os.FileMode(int(0660)))
	ohno(err)
}

func (f *FeedInfo) Save() {
	output, err := yaml.Marshal(f)
	ohno(err)

	// Markdown uses `---` for YAML frontmatter
	sep := []byte("---\n")
	output = append(sep, output...)
	output = append(output, sep...)

	path := "content/discover/feed-" + f.Params.FeedID + ".md"
	err = os.WriteFile(path, output, os.FileMode(int(0660)))
	ohno(err)
}

type ScanFeedInfo struct {
	Title       string
	Date        string
	Description string
	FeedLink    string
	FeedID      string
	FeedType    string
	IsPodcast   bool
	IsNoarchive bool
}

//type Link struct {
//	SourceType      int64
//	SourceURL       string
//	DestinationType int64
//	DestinationURL  string
//}

type LinkOnly struct {
	Link string
}

type TypedLink struct {
	Type int64
	Link string
}

type RelMeClusterInfo struct {
	Title  string             `yaml:"title"`
	Params RelMeClusterParams `yaml:"params"`
}

type RelMeClusterParams struct {
	VerifiedWebsites   []string            `yaml:"verifiedWebsites"`
	UnverifiedWebsites []string            `yaml:"unverifiedWebsites"`
	Feeds              map[string]FeedInfo `yaml:"feeds"`
}
