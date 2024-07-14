package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/looplab/tarjan"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Analysis struct {
	db               *gorm.DB
	relMeClusters    map[string]int
	relMeClustersRev map[int][]string
}

func NewAnalysis() *Analysis {
	var err error
	a := Analysis{}
	a.db, err = gorm.Open(sqlite.Open("feed2pages.db"), &gorm.Config{
		PrepareStmt: true,
	})
	ohno(err)
	return &a
}

func (a *Analysis) PopulateCategoriesForFeed(feed *FeedInfo) {
	var cats []string
	a.db.
		Model(&FeedsByCategory{}).
		Where("link = ?", feed.Params.FeedLink).
		Pluck("category", &cats)

	slices.Sort(cats)
	cats = slices.Compact(cats)
	feed.Params.Categories = append(feed.Params.Categories, cats...)

	if len(feed.Params.Categories) == 0 {
		a.PopulateCategoriesForFeedByHashtag(feed)
	}
}

func (a *Analysis) PopulateLanguageForFeed(feed *FeedInfo) {
	var language FeedsByLanguage
	result := a.db.
		Where("link = ?", feed.Params.FeedLink).
		First(&language)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return
	}
	ohno(result.Error)
	feed.Params.Language = language.Language
}

func (a *Analysis) PopulateCategoriesForPost(feed *FeedInfo) {
	post_link := feed.Params.LastPostLink
	if post_link == "" {
		return
	}

	var cats []string
	a.db.
		Model(&PostsByCategory{}).
		Where("link = ?", post_link).
		Pluck("category", &cats)

	slices.Sort(cats)
	cats = slices.Compact(cats)
	feed.Params.LastPostCategories = append(feed.Params.LastPostCategories, cats...)

	if len(feed.Params.LastPostCategories) == 0 {
		a.PopulateCategoriesForPostByHashtag(feed)
	}
}

func (a *Analysis) PopulateLanguageForPost(feed *FeedInfo) {
	post_link := feed.Params.LastPostLink
	if post_link == "" {
		return
	}

	var language PostsByLanguage
	result := a.db.
		Where("link = ?", post_link).
		First(&language)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return
	}
	ohno(result.Error)
	feed.Params.LastPostLanguage = language.Language
}

func isHashtag(s string) bool {
	// looks close enough
	// TODO: mirror the Twitter approach: https://github.com/twitter/twitter-text
	return strings.HasPrefix(s, "#") && len(s) > 2 && strings.LastIndex(s, "#") == 0
}

// Does the feed description have hashtags?
func (a *Analysis) PopulateCategoriesForFeedByHashtag(feed *FeedInfo) {
	// Description is already populated (when it exists)
	parts := strings.Fields(feed.Description)
	for _, part := range parts {
		if isHashtag(part) {
			feed.Params.Categories = append(feed.Params.Categories, part)
		}
	}
}

// Last resort, does the post description have hashtags?
func (a *Analysis) PopulateCategoriesForPostByHashtag(feed *FeedInfo) {
	// LastPostDesc is already populated (when it exists)
	parts := strings.Fields(feed.Params.LastPostDesc)
	for _, part := range parts {
		if isHashtag(part) {
			feed.Params.LastPostCategories = append(feed.Params.LastPostCategories, part)
		}
	}
}

func (a *Analysis) PopulateLastPostForFeed(feed *FeedInfo) {
	row := Post{}
	result := a.db.
		Where("feed_id = ?", feed.Params.FeedID).
		Order("date desc").
		First(&row)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return
	}
	ohno(result.Error)

	feed.Params.LastPostTitle = row.Title
	feed.Params.LastPostDesc = row.Description
	feed.Params.LastPostDate = row.Date
	feed.Params.LastPostLink = row.PostLink
	feed.Params.LastPostGuid = row.Guid

	a.PopulateCategoriesForPost(feed)
}

func (a *Analysis) PopulateBlogrollsForFeed(feed *FeedInfo) {
	// Feed => Blogroll or Website => Blogroll
	source_urls := []string{
		feed.Params.FeedLink,
	}
	for url, validated := range feed.Params.Websites {
		if validated {
			source_urls = append(source_urls, url)
		}
	}

	rows, err := a.db.
		Model(&Link{}).
		Select("destination_url").
		Where("destination_type = 3").
		Where("source_url IN(?)", source_urls).
		Rows()
	ohno(err)

	for rows.Next() {
		var row Link
		a.db.ScanRows(rows, &row)
		link := row.DestinationUrl
		if !slices.Contains(feed.Params.Blogrolls, link) {
			feed.Params.Blogrolls = append(feed.Params.Blogrolls, link)
		}
	}
}

// Website <==> Feed
func (a *Analysis) PopulateWebsitesForFeedURL(feed *FeedInfo) {
	rows, err := a.db.
		Model(&Link{}).
		Select("source_url").
		Where(Link{
			DestinationUrl:  feed.Params.FeedLink,
			DestinationType: NODE_TYPE_FEED,
			SourceType:      NODE_TYPE_WEBSITE,
		}).
		Rows()
	ohno(err)

	websites := []string{}
	for rows.Next() {
		var row Link
		a.db.ScanRows(rows, &row)
		link := row.SourceUrl
		if !slices.Contains(websites, link) {
			websites = append(websites, link)
			feed.Params.Websites[link] = false
		}
	}

	if len(websites) < 1 {
		return
	}

	// Now check that each feed points to the website
	// we want bidirectional links here to prevent
	// non-official pages from getting linked
	rows, err = a.db.
		Model(&Link{}).
		Select("destination_url").
		Where(Link{
			SourceUrl:       feed.Params.FeedLink,
			SourceType:      NODE_TYPE_FEED,
			DestinationType: NODE_TYPE_WEBSITE,
		}).
		Where("destination_url IN(?)", websites).
		Rows()
	ohno(err)

	for rows.Next() {
		var row Link
		a.db.ScanRows(rows, &row)
		link := row.DestinationUrl
		feed.Params.Websites[link] = true
	}
}

func (a *Analysis) PopulateRelMeForWebsites(feed *FeedInfo) {
	for url, validated := range feed.Params.Websites {
		if validated {
			clusterId, has := a.relMeClusters[url]
			if has {
				for _, link := range a.relMeClustersRev[clusterId] {
					feed.Params.RelMe[link] = true
				}
			}
		}
	}
}

func (a *Analysis) CollectWebsiteRecommendations(feed *FeedInfo) []string {
	if len(feed.Params.Blogrolls) < 1 {
		return []string{}
	}

	rows, err := a.db.
		Model(&Link{}).
		Select("destination_url", "destination_type").
		Where("source_url IN(?)", feed.Params.Blogrolls).
		Rows()
	ohno(err)

	websites := []string{}
	for rows.Next() {
		var row Link
		a.db.ScanRows(rows, &row)
		link := row.DestinationUrl
		linkType := row.DestinationType
		if linkType == NODE_TYPE_FEED {
			if !slices.Contains(feed.Params.Recommended, link) {
				feed.Params.Recommended = append(feed.Params.Recommended, link)
			}
		} else if linkType == NODE_TYPE_WEBSITE {
			if !slices.Contains(websites, link) {
				websites = append(websites, link)
			}
		}
	}
	return websites
}

func (a *Analysis) PopulateRecommendationsFromWebsites(feed *FeedInfo, websites []string) {
	if len(websites) < 1 {
		return
	}

	rows, err := a.db.
		Model(&Link{}).
		Select("destination_url").
		Where("destination_type = ?", NODE_TYPE_FEED).
		Where("source_url IN(?)", websites).
		Rows()
	ohno(err)

	fmt.Printf("\tDBG:Websites: %v\n", websites)
	for rows.Next() {
		var row Link
		a.db.ScanRows(rows, &row)
		link := row.DestinationUrl
		fmt.Printf("\tDBG: Feed From Websites: %s\n", link)

		if !slices.Contains(feed.Params.Recommended, link) {
			feed.Params.Recommended = append(feed.Params.Recommended, link)
		}
	}
}

func (a *Analysis) FindBlogrollsSuggestingFeed(feed *FeedInfo) []string {
	rows, err := a.db.
		Model(&Link{}).
		Select("source_url").
		Where(Link{
			DestinationUrl: feed.Params.FeedLink,
			SourceType:     NODE_TYPE_BLOGROLL,
		}).
		Rows()
	ohno(err)

	blogrolls := []string{}
	for rows.Next() {
		var row Link
		a.db.ScanRows(rows, &row)
		link := row.SourceUrl
		if !slices.Contains(blogrolls, link) {
			blogrolls = append(blogrolls, link)
		}
	}
	return blogrolls
}

func (a *Analysis) PopulateBlogrollPagesForFeed(feed *FeedInfo, blogrolls []string, blogrollMap map[string]*BlogrollInfo) {
	for _, blogroll := range blogrolls {
		data, has := blogrollMap[blogroll]
		if has {
			feed.Params.InBlogroll = append(
				feed.Params.InBlogroll,
				BlogrollBrief{
					Title:       data.Title,
					Description: data.Description,
					Id:          data.Params.BlogrollId,
				},
			)
		}
	}
}

func (a *Analysis) PopulateFeedForBlogroll(feed *FeedInfo, blogrolls []string, blogrollMap map[string]*BlogrollInfo) {
	for _, blogroll := range blogrolls {
		data, has := blogrollMap[blogroll]
		if has {
			data.Params.Recommends = append(data.Params.Recommends,
				BlogrollRecommendation{
					Title:       feed.Title,
					Description: feed.Description,
					Id:          feed.Params.FeedID,
				})
		}
	}
}

func (a *Analysis) FindWebsitesRecommendingBlogrolls(blogrolls []string) []string {
	if len(blogrolls) < 1 {
		return []string{}
	}

	rows, err := a.db.
		Model(&Link{}).
		Select("source_url").
		Where("source_type = ?", NODE_TYPE_WEBSITE).
		Where("destination_url IN(?)", blogrolls).
		Rows()
	ohno(err)

	websites := []string{}
	for rows.Next() {
		var row Link
		a.db.ScanRows(rows, &row)
		link := row.SourceUrl
		if !slices.Contains(websites, link) {
			websites = append(websites, link)
		}
	}
	return websites
}

func (a *Analysis) PopulateScore(feed *FeedInfo) {
	isSocial := false

	// Does this site recommend others?
	// More recommendations are better
	// until you reach 20
	// Half a point each, up to 10 points
	promotesScore := min(len(feed.Params.Recommended), 20) / 2
	feed.Params.ScoreCriteria["promotes"] = promotesScore
	if promotesScore > 0 {
		isSocial = true
	}

	// Do others recommend this feed?
	// 5 points if any
	promotedScore := 0
	if len(feed.Params.InBlogroll) > 0 {
		promotedScore = 5
	}
	feed.Params.ScoreCriteria["promoted"] = promotedScore
	if promotedScore > 0 {
		isSocial = true
	}

	// Does this feed have a website?
	// +1 point
	// Can we verify the website via backlinks?
	// +1 point
	websiteScore := 0
	for _, verified := range feed.Params.Websites {
		if verified {
			websiteScore = 2
		} else {
			websiteScore = max(websiteScore, 1)
		}
	}
	feed.Params.ScoreCriteria["website"] = websiteScore

	// Is this page related to others (rel=me)?
	// +1 point
	// Are they verified?
	// +1 point
	relMeScore := 0
	for _, verified := range feed.Params.RelMe {
		if verified {
			relMeScore = 2
			//isSocial = true // TODO: keep this?
		} else {
			relMeScore = max(relMeScore, 1)
		}
	}
	feed.Params.ScoreCriteria["relme"] = relMeScore

	// Does the feed have categories?
	// One point each, up to five
	catScore := min(len(feed.Params.Categories), 5)
	feed.Params.ScoreCriteria["cats"] = catScore

	// Does the last post have categories?
	// One point each, up to three
	postCatScore := min(len(feed.Params.LastPostCategories), 3)
	feed.Params.ScoreCriteria["postcats"] = postCatScore

	// Does the feed specify a language?
	feedLangScore := min(len(feed.Params.Language), 1)
	feed.Params.ScoreCriteria["feedlangs"] = feedLangScore

	// Does the feed have a title?
	// 3 points
	titleScore := 0
	if len(feed.Title) > 0 {
		titleScore = 3
	}
	feed.Params.ScoreCriteria["title"] = titleScore

	// Does the feed have a description?
	// 3 points
	descriptionScore := 0
	if len(feed.Description) > 0 {
		descriptionScore = 3
	}
	feed.Params.ScoreCriteria["description"] = descriptionScore

	// Sum it up
	feed.Params.Score = 0
	for _, score := range feed.Params.ScoreCriteria {
		feed.Params.Score += score
	}

	// Track if the feed is part of the network
	feed.Params.InNetwork = isSocial
}

func (a *Analysis) PopulateRecommenders(feed *FeedInfo, blogrolls []string, websites []string) {
	targetUrls := append(blogrolls, websites...)

	if len(targetUrls) < 1 {
		return
	}

	rows, err := a.db.
		Model(&Link{}).
		Select("source_url").
		Where("source_type = ?", NODE_TYPE_FEED).
		Where("destination_url IN(?)", targetUrls).
		Rows()
	ohno(err)

	for rows.Next() {
		var row Link
		a.db.ScanRows(rows, &row)
		link := row.SourceUrl
		if !slices.Contains(feed.Params.Recommender, link) {
			feed.Params.Recommender = append(feed.Params.Recommender, link)
		}
	}
}

func (a *Analysis) FixUp(feed *FeedInfo) {
	// Hack: when the feed doesn't have categories, but the last post does
	// use the post categories as the feed categories
	if len(feed.Params.Categories) == 0 {
		feed.Params.Categories = feed.Params.LastPostCategories
	}

	// Assume the feed generally uses the language of the most recent post
	if len(feed.Params.Language) == 0 {
		feed.Params.Language = feed.Params.LastPostLanguage
	}

	// Something in hugo breaks if there's a trailing slash
	for i, c := range feed.Params.Categories {
		if strings.HasSuffix(c, "/") {
			feed.Params.Categories[i] = c + " "
		}
	}
	for i, c := range feed.Params.LastPostCategories {
		if strings.HasSuffix(c, "/") {
			feed.Params.LastPostCategories[i] = c + " "
		}
	}
}

func (a *Analysis) BuildRelMeClusters() (map[string]int, map[int][]string) {
	links := make(map[interface{}][]interface{})
	rows, err := a.db.
		Model(&Link{}).
		Select("source_url", "destination_url").
		Where("link_type = ?", "rel_me").
		Rows()
	ohno(err)

	for rows.Next() {
		var row Link
		a.db.ScanRows(rows, &row)
		source_url := row.SourceUrl
		destination_url := row.DestinationUrl
		if _, has := links[source_url]; !has {
			links[source_url] = []interface{}{}
		}
		links[source_url] = append(links[source_url], destination_url)
	}

	// Find strongly connected components
	// These are verified rel=me links
	connections := tarjan.Connections(links)

	// Restructure
	out := map[string]int{}
	outRev := map[int][]string{}
	connId := 0
	for _, connected := range connections {
		group := []string{}
		for _, vertex := range connected {
			out[vertex.(string)] = connId
			group = append(group, vertex.(string))
		}
		outRev[connId] = group
		connId += 1
	}

	return out, outRev
}

func (a *Analysis) Analyze() {
	rows, err := a.db.
		Model(&Blogroll{}).
		Rows()
	ohno(err)
	blogrollDataByLink := map[string]*BlogrollInfo{}
	for rows.Next() {
		var row Blogroll
		a.db.ScanRows(rows, &row)
		blogroll := NewBlogrollInfo(row)
		blogrollDataByLink[blogroll.Params.Link] = blogroll
	}

	// Tarjan to consolidate all verified rel=me profiles
	a.relMeClusters, a.relMeClustersRev = a.BuildRelMeClusters()
	// clusters - { A=>1, B=>1, C=>2, D=2, E=>3, F=4 }
	// reverse  - { 1=>[A,B], 2=>[C,D], 3=>[E], 4=>[F] }

	rows, err = a.db.
		Model(&Feed{}).
		Rows()
	ohno(err)
	for rows.Next() {
		var row Feed
		a.db.ScanRows(rows, &row)
		feed := NewFeedInfo(row)
		fmt.Printf("\n\nProcessing Feed: %s\n", feed.Title)
		a.PopulateWebsitesForFeedURL(feed)
		a.PopulateBlogrollsForFeed(feed)
		fmt.Printf("\tOut Blogrolls: %v\n", feed.Params.Blogrolls)
		out_websites := a.CollectWebsiteRecommendations(feed)
		fmt.Printf("\tOut Websites: %v\n", out_websites)
		a.PopulateRecommendationsFromWebsites(feed, out_websites)
		fmt.Printf("\tOut Recommended: %v\n", feed.Params.Recommended)
		rec_blogrolls := a.FindBlogrollsSuggestingFeed(feed)
		fmt.Printf("\tIn Blogrolls: %v\n", rec_blogrolls)
		rec_websites := a.FindWebsitesRecommendingBlogrolls(rec_blogrolls)
		fmt.Printf("\tIn Websites: %v\n", rec_websites)
		a.PopulateRecommenders(feed, rec_blogrolls, rec_websites)
		fmt.Printf("\tIn Feeds: %v\n", feed.Params.Recommender)
		a.PopulateRelMeForWebsites(feed)
		a.PopulateCategoriesForFeed(feed)
		a.PopulateLanguageForFeed(feed)
		a.PopulateLastPostForFeed(feed)
		a.PopulateBlogrollPagesForFeed(feed, rec_blogrolls, blogrollDataByLink)
		a.PopulateFeedForBlogroll(feed, rec_blogrolls, blogrollDataByLink)

		a.PopulateScore(feed)

		// Apply some hacks to improve content
		// but do this after the score is calculated
		a.FixUp(feed)

		// Ignore feeds outside the network or without content
		if feed.Params.InNetwork && len(feed.Params.LastPostTitle) > 0 {
			feed.Save()
		}
	}

	for _, data := range blogrollDataByLink {
		data.Save()
	}
}
