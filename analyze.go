package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Analysis struct {
	db *sqlx.DB
}

func NewAnalysis() *Analysis {
	var err error
	a := Analysis{}
	a.db, err = sqlx.Open("sqlite3", "feed2pages.db")
	ohno(err)
	err = a.db.Ping()
	ohno(err)
	return &a
}

func (a *Analysis) Close() {
	a.db.Close()
}

func (a *Analysis) PopulateCategoriesForFeed(feed *FeedInfo) {
	rows, err := a.db.Query(`
    SELECT category
      FROM feeds_by_categories
     WHERE link = ?
    `,
		feed.Params.FeedLink,
	)
	ohno(err)

	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)
		if !slices.Contains(feed.Params.Categories, link) {
			feed.Params.Categories = append(feed.Params.Categories, link)
		}
	}

	if len(feed.Params.Categories) == 0 {
		a.PopulateCategoriesForFeedByHashtag(feed)
	}
}

func (a *Analysis) PopulateCategoriesForPost(feed *FeedInfo) {
	post_link := feed.Params.LastPostLink
	if post_link == "" {
		return
	}
	rows, err := a.db.Query(`
    SELECT category
      FROM posts_by_categories
     WHERE link = ?
    `,
		post_link,
	)
	ohno(err)

	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)
		if !slices.Contains(feed.Params.LastPostCategories, link) {
			feed.Params.LastPostCategories = append(feed.Params.LastPostCategories, link)
		}
	}

	if len(feed.Params.LastPostCategories) == 0 {
		a.PopulateCategoriesForPostByHashtag(feed)
	}
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
	rows, err := a.db.Query(`
       SELECT date, description, title, post_link
         FROM posts
        WHERE feed_id = ?
     ORDER BY date DESC
        LIMIT 1
    `,
		feed.Params.FeedID,
	)
	ohno(err)

	var date string
	var description string
	var title string
	var link string
	for rows.Next() {
		err = rows.Scan(&date, &description, &title, &link)
		ohno(err)
	}
	feed.Params.LastPostTitle = title
	feed.Params.LastPostDesc = description
	feed.Params.LastPostDate = date
	feed.Params.LastPostLink = link

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

	query, args, err := sqlx.In(`
    SELECT destination_url
      FROM links
     WHERE destination_type = 3
       AND source_url IN(?);
    `,
		source_urls,
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err := a.db.Query(query, args...)
	ohno(err)

	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)
		if !slices.Contains(feed.Params.Blogrolls, link) {
			feed.Params.Blogrolls = append(feed.Params.Blogrolls, link)
		}
	}
}

// Website <==> Feed
func (a *Analysis) PopulateWebsitesForFeedURL(feed *FeedInfo) {
	// Find websites that point to this feed
	rows, err := a.db.Query(`
    SELECT source_url
      FROM links
     WHERE destination_url = ?
       AND destination_type = ?
       AND source_type = ?
    `,
		feed.Params.FeedLink,
		NODE_TYPE_FEED,
		NODE_TYPE_WEBSITE,
	)
	ohno(err)

	websites := []string{}
	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)
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
	query, args, err := sqlx.In(`
    SELECT destination_url
      FROM links
     WHERE source_url = ?
       AND source_type = ?
       AND destination_url IN(?)
       AND destination_type = ?
    `,
		feed.Params.FeedLink,
		NODE_TYPE_FEED,
		websites,
		NODE_TYPE_WEBSITE,
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err = a.db.Query(
		query,
		args...,
	)
	ohno(err)

	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)
		feed.Params.Websites[link] = true
	}
}

func (a *Analysis) PopulateRelMeForWebsites(feed *FeedInfo) {
	// Only consider rel=me links from validated pages
	validatedWebsites := []string{}
	for url, validated := range feed.Params.Websites {
		if validated {
			validatedWebsites = append(validatedWebsites, url)
		}
	}
	if len(validatedWebsites) < 1 {
		return
	}

	query, args, err := sqlx.In(`
    SELECT destination_url
      FROM links
     WHERE source_url IN(?)
       AND link_type = ?
    `,
		validatedWebsites,
		"rel_me",
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err := a.db.Query(
		query,
		args...,
	)
	ohno(err)

	pendingRelMe := []string{}
	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)
		pendingRelMe = append(pendingRelMe, link)
		feed.Params.RelMe[link] = false
	}

	if len(pendingRelMe) < 1 {
		return
	}

	// Look for backlinks
	query, args, err = sqlx.In(`
    SELECT source_url
      FROM links
     WHERE source_url IN(?)
       AND destination_url IN(?)
       AND link_type = ?
    `,
		pendingRelMe,
		validatedWebsites,
		"rel_me",
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err = a.db.Query(
		query,
		args...,
	)
	ohno(err)

	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)
		// Validated!
		feed.Params.RelMe[link] = true
	}
}

func (a *Analysis) CollectWebsiteRecommendations(feed *FeedInfo) []string {
	if len(feed.Params.Blogrolls) < 1 {
		return []string{}
	}

	query, args, err := sqlx.In(`
    SELECT destination_url, destination_type
      FROM links
     WHERE source_url IN(?);
    `,
		feed.Params.Blogrolls,
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err := a.db.Query(
		query,
		args...,
	)
	ohno(err)

	websites := []string{}
	for rows.Next() {
		var link string
		var linkType int64
		err = rows.Scan(&link, &linkType)
		ohno(err)
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

	query, args, err := sqlx.In(`
    SELECT destination_url
      FROM links
     WHERE destination_type = ?
       AND source_url IN(?);
    `,
		NODE_TYPE_FEED,
		websites,
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err := a.db.Query(
		query,
		args...,
	)
	ohno(err)

	fmt.Printf("\tDBG:Websites: %v\n", websites)
	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)

		fmt.Printf("\tDBG: Feed From Websites: %s\n", link)

		if !slices.Contains(feed.Params.Recommended, link) {
			feed.Params.Recommended = append(feed.Params.Recommended, link)
		}
	}
}

func (a *Analysis) FindBlogrollsSuggestingFeed(feed *FeedInfo) []string {
	query, args, err := sqlx.In(`
    SELECT source_url
      FROM links
     WHERE destination_url = ?
       AND source_type = ?;
    `,
		feed.Params.FeedLink,
		NODE_TYPE_BLOGROLL,
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err := a.db.Query(
		query,
		args...,
	)
	ohno(err)

	blogrolls := []string{}
	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)

		if !slices.Contains(blogrolls, link) {
			blogrolls = append(blogrolls, link)
		}
	}
	return blogrolls
}

func (a *Analysis) FindWebsitesRecommendingBlogrolls(blogrolls []string) []string {
	if len(blogrolls) < 1 {
		return []string{}
	}

	query, args, err := sqlx.In(`
    SELECT source_url
      FROM links
     WHERE destination_url IN(?)
       AND source_type = ?;
    `,
		blogrolls,
		NODE_TYPE_WEBSITE,
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err := a.db.Query(
		query,
		args...,
	)
	ohno(err)

	websites := []string{}
	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)

		if !slices.Contains(websites, link) {
			websites = append(websites, link)
		}
	}
	return websites
}

func (a *Analysis) PopulateScore(feed *FeedInfo) {
	// Does this site recommend others?
	// More recommendations are better
	// until you reach 20
	// Half a point each, up to 10 points
	promotesScore := min(len(feed.Params.Recommended), 20) / 2
	feed.Params.ScoreCriteria["promotes"] = promotesScore

	// Do others recommend this feed?
	// 5 points if any
	promotedScore := 0
	if len(feed.Params.Recommender) > 0 {
		promotedScore = 5
	}
	feed.Params.ScoreCriteria["promoted"] = promotedScore

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
}

func (a *Analysis) PopulateRecommenders(feed *FeedInfo, blogrolls []string, websites []string) {
	targetUrls := append(blogrolls, websites...)

	if len(targetUrls) < 1 {
		return
	}

	query, args, err := sqlx.In(`
    SELECT source_url
      FROM links
     WHERE destination_url IN(?)
       AND source_type = ?;
    `,
		targetUrls,
		NODE_TYPE_FEED,
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err := a.db.Query(
		query,
		args...,
	)
	ohno(err)

	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)

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

func (a *Analysis) Analyze() {
	feedRows, err := a.db.Query(`
    SELECT description, date, title, feed_link, feed_id, feed_type, is_podcast
      FROM feeds;`,
	)
	ohno(err)
	for feedRows.Next() {
		var row ScanFeedInfo
		err = feedRows.Scan(
			&row.Description,
			&row.Date,
			&row.Title,
			&row.FeedLink,
			&row.FeedID,
			&row.FeedType,
			&row.IsPodcast,
		)
		ohno(err)
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
		a.PopulateLastPostForFeed(feed)
		a.PopulateScore(feed)

		// Apply some hacks to improve content
		// but do this after the score is calculated
		a.FixUp(feed)

		feed.Save()
	}
}
