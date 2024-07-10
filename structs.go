package main

import (
	"github.com/go-yaml/yaml"
	"os"
)

type FeedInfo struct {
	Title       string         `yaml:"title"`
	Date        string         `yaml:"date"`
	Description string         `yaml:"description"`
	Params      FeedInfoParams `yaml:"params"`
}

type FeedInfoParams struct {
	FeedLink           string          `yaml:"feedlink"`
	FeedType           string          `yaml:"feedtype"`
	FeedID             string          `yaml:"feedid"`
	Websites           map[string]bool `yaml:"websites"`
	Blogrolls          []string        `yaml:"blogrolls"`
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
