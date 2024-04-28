package main

import (
	"errors"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/mmcdole/gofeed"
	"os"
	"strings"
	"time"
)

const POST_FOLDER_PATH = "content/post"
const FEED_FOLDER_PATH = "content/feed"

func filterPost(item *MultiTypeItem, feedConfig FeedConfig, config Config) error {
	// Missing required fields
	if item.item.PublishedParsed == nil {
		return errMissingField("PublishedParsed")
	}
	if item.item.Link == "" {
		return errMissingField("Link")
	}
	if item.item.Title == "" {
		return errMissingField("Title")
	}

	// Filter out banned words
	if has, which := containsAny(item.item.Description, config.BlockWords...); has {
		return errBlockWord("Description", which)
	}
	if has, which := containsAny(item.item.Title, config.BlockWords...); has {
		return errBlockWord("Title", which)
	}
	if has, which := containsAny(item.item.Content, config.BlockWords...); has {
		return errBlockWord("Content", which)
	}
	if has, which := containsAny(item.item.Description, feedConfig.BlockWords...); has {
		return errBlockWord("Description", which)
	}
	if has, which := containsAny(item.item.Title, feedConfig.BlockWords...); has {
		return errBlockWord("Title", which)
	}
	if has, which := containsAny(item.item.Content, feedConfig.BlockWords...); has {
		return errBlockWord("Content", which)
	}

	// Blocked domains
	for _, blockedDomain := range config.BlockDomains {
		if isDomainOrSubdomain(item.item.Link, blockedDomain) {
			return errors.New(fmt.Sprintf("Domain is blocked: %s", blockedDomain))

		}
	}
	for _, blockedDomain := range feedConfig.BlockDomains {
		if isDomainOrSubdomain(item.item.Link, blockedDomain) {
			return errors.New(fmt.Sprintf("Domain is blocked: %s", blockedDomain))

		}
	}

	return nil
}

func processPost(item *MultiTypeItem, feed *gofeed.Feed, feedConfig FeedConfig, config Config) (PostFrontmatter, error) {
	out := PostFrontmatter{}
	out.Params.Feed = *feed
	out.Params.Feed.Items = []*gofeed.Item{} // exclude the others posts
	out.Params.Post = *item.item
	out.Date = item.item.PublishedParsed.Format(time.RFC3339)

	// Filter out content that's too old
	age := time.Since(*item.item.PublishedParsed)
	out.Params.PrettyAge = pretty(age)
	ageDays := int(age.Hours() / 24)
	if config.PostAgeLimitDays != nil && ageDays > *config.PostAgeLimitDays {
		return out, errors.New("Too old")
	}

	if feedConfig.IgnoreDescription != nil && *feedConfig.IgnoreDescription {
		out.Params.Post.Description = ""
	}
	if feedConfig.IgnoreContent != nil && *feedConfig.IgnoreContent {
		out.Params.Post.Content = ""
	}

	// The description is one of:
	//  - description from the feed
	//  - the content from the feed
	//  - the contents of the linked page (HTTP GET it)
	// Pass content through readability to remove HTML and cruft
	isUsingContentAsDescription := false
	if isEmpty(out.Params.Post.Description) {
		isUsingContentAsDescription = true
		out.Params.Post.Description = out.Params.Post.Content
	}
	out.Params.Post.Description = readable(out.Params.Post.Description)
	if isEmpty(out.Params.Post.Description) {
		isUsingContentAsDescription = true
		out.Params.Post.Description = readablePost(out.Params.Post.Link)
	}
	out.Params.Post.Description = strings.TrimSpace(out.Params.Post.Description)

	err := filterPost(item, feedConfig, config)
	if err != nil {
		return out, err
	}

	// Ensure our summary isn't too much of the content
	if isUsingContentAsDescription {
		out.Params.Post.Description = truncateText(
			out.Params.Post.Description,
			len(out.Params.Post.Description)/20,
		)
	}

	// An RSS only field (not Atom or JSON feeds)
	if item.rss != nil {
		out.Params.CommentsLink = item.rss.Comments
	}

	// Reduce the size, we won't render it all anyway
	out.Title = truncateText(readable(item.item.Title), 200)
	out.Params.Post.Description = truncateText(out.Params.Post.Description, 1024)
	out.Params.Post.Content = truncateText(out.Params.Post.Content, 1024)

	return out, nil
}

func processFeed(feedId string, feedConfig FeedConfig, config Config) ([]PostFrontmatter, FeedFrontmatter) {
	fp := NewParser()
	parsedFeed, mergedItems, err := fp.ParseURLExtended(feedConfig.URL)
	if err != nil {
		fmt.Printf("Unable to parse feed: %s %v", feedConfig.URL, err)
		return []PostFrontmatter{}, FeedFrontmatter{}
	}

	feed := FeedFrontmatter{}
	feed.Title = parsedFeed.Title
	feed.Description = parsedFeed.Description
	feed.Params.Feed = *parsedFeed
	feed.Params.Id = feedId

	// Store posts elsewhere
	feed.Params.Feed.Items = []*gofeed.Item{}

	if parsedFeed.PublishedParsed != nil {
		feed.Date = parsedFeed.PublishedParsed.Format(time.RFC3339)
	} else if parsedFeed.UpdatedParsed != nil {
		feed.Date = parsedFeed.UpdatedParsed.Format(time.RFC3339)
	}

	posts := []PostFrontmatter{}
	for _, post := range mergedItems {
		processed, err := processPost(&post, parsedFeed, feedConfig, config)
		processed.Params.FeedId = feedId
		if err != nil {
			fmt.Printf("  Excluding post: %v\n", err)
			continue
		}
		posts = append(posts, processed)
	}
	return posts, feed
}

func writePost(post PostFrontmatter) {
	output, err := yaml.Marshal(post)
	if err != nil {
		panic(fmt.Sprintf("YAML error: %v\n", err))
	}

	// Markdown uses `---` for YAML frontmatter
	sep := []byte("---\n")
	output = append(sep, output...)
	output = append(output, sep...)

	path := fmt.Sprintf("%s/%s.md", POST_FOLDER_PATH, safeGUID(post))
	err = os.WriteFile(path, output, os.FileMode(int(0600)))
	if err != nil {
		panic(fmt.Sprintf("Unable to write file %s %v", path, err))
	}
}

func writeFeed(feed FeedFrontmatter) {
	output, err := yaml.Marshal(feed)
	if err != nil {
		panic(fmt.Sprintf("YAML error: %v\n", err))
	}

	// Markdown uses `---` for YAML frontmatter
	sep := []byte("---\n")
	output = append(sep, output...)
	output = append(output, sep...)

	path := fmt.Sprintf("%s/%s.md", FEED_FOLDER_PATH, feed.Params.Id)
	err = os.WriteFile(path, output, os.FileMode(int(0600)))
	if err != nil {
		panic(fmt.Sprintf("Unable to write file %s %v", path, err))
	}
}

func main() {
	rmdir(POST_FOLDER_PATH)
	rmdir(FEED_FOLDER_PATH)
	mkdirIfNotExists(POST_FOLDER_PATH)
	mkdirIfNotExists(FEED_FOLDER_PATH)
	config := parseConfig()
	allPosts := []PostFrontmatter{}
	allFeeds := []FeedFrontmatter{}
	for id, feedConfig := range config.Feeds {
		fmt.Printf("Processing feed: %s\n", feedConfig.URL)
		feedId := fmt.Sprintf("%d", id)
		posts, feed := processFeed(feedId, feedConfig, config)
		allFeeds = append(allFeeds, feed)
		posts = sortAndLimitPosts(posts, *config.MaxPostsPerFeed)
		fmt.Printf("  got %d posts\n", len(posts))
		allPosts = append(allPosts, posts...)
	}
	allPosts = sortAndLimitPosts(allPosts, *config.MaxPosts)
	fmt.Printf("Total %d posts\n", len(allPosts))
	for _, feed := range allFeeds {
		writeFeed(feed)
	}
	for _, post := range allPosts {
		writePost(post)
	}
}
