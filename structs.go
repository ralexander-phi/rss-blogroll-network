package main

import (
	"github.com/go-yaml/yaml"
	"os"
)

type FeedInfo struct {
	Title       string         `yaml:"title"`
	Description string         `yaml:"description"`
	Params      FeedInfoParams `yaml:"params"`
}

type FeedInfoParams struct {
	FeedLink    string          `yaml:"feedlink"`
	FeedType    string          `yaml:"feedtype"`
	FeedID      string          `yaml:"feedid"`
	Websites    map[string]bool `yaml:"websites"`
	Blogrolls   []string        `yaml:"blogrolls"`
	Recommended []string        `yaml:"recommended"`
	Recommender []string        `yaml:"recommender"`
	Categories  []string        `yaml:"categories"`
	RelMe       map[string]bool `yaml:"relme"`
}

func NewFeedInfo(row ScanFeedInfo) *FeedInfo {
	p := FeedInfoParams{
		FeedLink: row.FeedLink,
		FeedID:   row.FeedID,
		FeedType: row.FeedType,
		Websites: map[string]bool{},
		RelMe:    map[string]bool{},
	}
	f := FeedInfo{
		Title:       row.Title,
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
	Description string
	FeedLink    string
	FeedID      string
	FeedType    string
}

type Link struct {
	SourceType      int64
	SourceURL       string
	DestinationType int64
	DestinationURL  string
}

type LinkOnly struct {
	Link string
}

type TypedLink struct {
	Type int64
	Link string
}
