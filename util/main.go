package main

import (
	"errors"
	"fmt"
	"github.com/mmcdole/gofeed"
	"time"
)

const POST_FOLDER_PATH = "content/post"
const FEED_FOLDER_PATH = "content/feed"

func filterPost(item *MultiTypeItem, config Config) error {
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

	// Blocked domains
	for _, blockedDomain := range config.BlockDomains {
		if isDomainOrSubdomain(item.item.Link, blockedDomain) {
			return errors.New(fmt.Sprintf("Domain is blocked: %s", blockedDomain))
		}
	}
	return nil
}

func processPost(item *MultiTypeItem, feed *gofeed.Feed, config Config) (PostFrontmatter, error) {
	out := PostFrontmatter{}
	out.Params.Feed = *feed
	out.Params.Feed.Items = []*gofeed.Item{} // exclude the others posts
	out.Params.Post = *item.item

	postDate := unixEpoc()
	if item.item.PublishedParsed != nil {
		postDate = *item.item.PublishedParsed
	}
	out.Date = postDate.Format(time.RFC3339)
	age := time.Since(postDate)
	out.Params.PrettyAge = pretty(age)
	ageDays := int(age.Hours() / 24)

	// Filter out content that's too old
	if config.PostAgeLimitDays != nil && ageDays > *config.PostAgeLimitDays {
		return out, errors.New("Too old")
	}

	// The description is one of:
	//  - description from the feed
	//  - the content from the feed
	out.Params.Post.Description = firstNonEmpty(
		[]string{
			out.Params.Post.Description,
			out.Params.Post.Content,
		})

	err := filterPost(item, config)
	if err != nil {
		return out, err
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

func processFeed(feedId string, feedDetails FeedDetails, config Config) ([]PostFrontmatter, FeedFrontmatter) {
	fp := NewParser()
	parsedFeed, mergedItems, err := fp.ParseURLExtended(feedDetails.Link)
	if err != nil {
		fmt.Printf("Unable to parse feed: %v %v", feedDetails, err)
		return []PostFrontmatter{}, FeedFrontmatter{}
	}

	feed := FeedFrontmatter{}
	feed.Title = parsedFeed.Title
	feed.Description = parsedFeed.Description
	feed.Params.Feed = *parsedFeed
	feed.Params.Id = feedId

	if isEmpty(feed.Title) {
		feed.Title = feedDetails.Title
	}
	if isEmpty(feed.Description) {
		feed.Description = feedDetails.Text
	}

	// Store posts elsewhere
	feed.Params.Feed.Items = []*gofeed.Item{}

	if parsedFeed.PublishedParsed != nil {
		feed.Date = parsedFeed.PublishedParsed.Format(time.RFC3339)
	} else if parsedFeed.UpdatedParsed != nil {
		feed.Date = parsedFeed.UpdatedParsed.Format(time.RFC3339)
	}

	posts := []PostFrontmatter{}
	for _, post := range mergedItems {
		processed, err := processPost(&post, parsedFeed, config)
		processed.Params.FeedId = feedId
		if err != nil {
			fmt.Printf("  Excluding post: %v\n", err)
			continue
		}
		posts = append(posts, processed)
	}
	return posts, feed
}

func main() {
	rmdir(POST_FOLDER_PATH)
	rmdir(FEED_FOLDER_PATH)
	mkdirIfNotExists(POST_FOLDER_PATH)
	mkdirIfNotExists(FEED_FOLDER_PATH)
	config := parseConfig()
	allPosts := []PostFrontmatter{}
	allFeeds := []FeedFrontmatter{}
	for id, feedDetails := range config.Feeds {
		fmt.Printf("Processing feed: %v\n", feedDetails)
		feedId := fmt.Sprintf("%d", id)
		posts, feed := processFeed(feedId, feedDetails, config)
		allFeeds = append(allFeeds, feed)
		posts = sortAndLimitPosts(posts, *config.MaxPostsPerFeed)
		fmt.Printf("  got %d posts\n", len(posts))
		allPosts = append(allPosts, posts...)
	}
	allPosts = sortAndLimitPosts(allPosts, *config.MaxPosts)
	fmt.Printf("Total %d posts\n", len(allPosts))
	for _, feed := range allFeeds {
		path := fmt.Sprintf("%s/%s.md", FEED_FOLDER_PATH, feed.Params.Id)
		writeYaml(feed, path)
	}
	for _, post := range allPosts {
		path := fmt.Sprintf("%s/%s.md", POST_FOLDER_PATH, safeGUID(post))
		writeYaml(post, path)
	}
}
