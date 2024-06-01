package main

import (
	"fmt"
	"slices"

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
	query, args, err := sqlx.In(`
    SELECT category
      FROM feeds_by_categories
     WHERE link = ?
    `,
		feed.Params.FeedLink,
	)
	ohno(err)

	query = a.db.Rebind(query)
	rows, err := a.db.Query(query, args...)
	ohno(err)

	for rows.Next() {
		var link string
		err = rows.Scan(&link)
		ohno(err)
		if !slices.Contains(feed.Params.Categories, link) {
			feed.Params.Categories = append(feed.Params.Categories, link)
		}
	}
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

func (a *Analysis) Analyze() {
	feedRows, err := a.db.Query(`
    SELECT description, title, feed_link, feed_id, feed_type
      FROM feeds;`,
	)
	ohno(err)
	for feedRows.Next() {
		var row ScanFeedInfo
		err = feedRows.Scan(
			&row.Description,
			&row.Title,
			&row.FeedLink,
			&row.FeedID,
			&row.FeedType,
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
		feed.Save()
	}
}
