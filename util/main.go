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

const POST_FOLDER_PATH = "content/posts"

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

func processPost(item *MultiTypeItem, feedConfig FeedConfig, config Config) (Frontmatter, error) {
	out := Frontmatter{}
	err := filterPost(item, feedConfig, config)
	if err != nil {
		return out, err
	}

	// Filter out content that's too old
	age := time.Since(*item.item.PublishedParsed)
	ageDays := int(age.Hours() / 24)
	if config.PostAgeLimitDays != nil && ageDays > *config.PostAgeLimitDays {
		return out, errors.New("Too old")
	}

	out.Params.Feed = *item.feed
	out.Params.Feed.Items = []*gofeed.Item{} // exclude the others posts
	out.Params.Post = *item.item

	out.Title = truncateText(readable(item.item.Title), 200)
	out.Date = item.item.PublishedParsed.Format(time.RFC3339)
	out.Params.PrettyAge = pretty(age)

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
	if isEmpty(out.Params.Post.Description) {
		out.Params.Post.Description = out.Params.Post.Content
	}
	out.Params.Post.Description = readable(out.Params.Post.Description)
	if isEmpty(out.Params.Post.Description) {
		out.Params.Post.Description = readablePost(out.Params.Post.Link)
	}
	out.Params.Post.Description = strings.TrimSpace(out.Params.Post.Description)

	// An RSS only field (not Atom or JSON feeds)
	if item.rss != nil {
		out.Params.CommentsLink = item.rss.Comments
	}

	// Reduce the size, we won't render it all anyway
	out.Params.Post.Description = truncateText(out.Params.Post.Description, 1024)
	out.Params.Post.Content = truncateText(out.Params.Post.Content, 1024)

	return out, nil
}

func processFeed(feedConfig FeedConfig, config Config) []Frontmatter {
	fp := NewParser()
	mergedItems, err := fp.ParseURLExtended(feedConfig.URL)
	if err != nil {
		fmt.Printf("Unable to parse feed: %s %v", feedConfig.URL, err)
		return []Frontmatter{}
	}

	posts := []Frontmatter{}
	for _, post := range mergedItems {
		processed, err := processPost(&post, feedConfig, config)
		if err != nil {
			fmt.Printf("  Excluding post: %v\n", err)
			continue
		}
		posts = append(posts, processed)
	}
	return posts
}

func writePost(post Frontmatter) {
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

func main() {
	rmdir(POST_FOLDER_PATH)
	mkdirIfNotExists(POST_FOLDER_PATH)
	config := parseConfig()
	allPosts := []Frontmatter{}
	for _, feedConfig := range config.Feeds {
		fmt.Printf("Processing feed: %s\n", feedConfig.URL)
		posts := processFeed(feedConfig, config)
		posts = sortAndLimitPosts(posts, *config.MaxPostsPerFeed)
		fmt.Printf("  got %d posts\n", len(posts))
		allPosts = append(allPosts, posts...)
	}
	allPosts = sortAndLimitPosts(allPosts, *config.MaxPosts)
	fmt.Printf("Total %d posts\n", len(allPosts))
	for _, post := range allPosts {
		writePost(post)
	}
}
