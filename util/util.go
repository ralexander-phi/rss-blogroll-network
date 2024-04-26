package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	readability "github.com/go-shiori/go-readability"
	"github.com/go-yaml/yaml"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

func anyContains(target string, searchAmong ...string) (bool, string) {
	for _, search := range searchAmong {
		if strings.Contains(search, target) {
			return true, search
		}
	}
	return false, ""
}

func containsAny(search string, targets ...string) (bool, string) {
  // Case insensitive
  search = strings.ToLower(search)
	for _, target := range targets {
		if strings.Contains(search, strings.ToLower(target)) {
			return true, target
		}
	}
	return false, ""
}

func mkdirIfNotExists(path string) {
	err := os.MkdirAll(path, 0660)
	if err != nil {
		panic(fmt.Sprintf("Unable to creater directory: %s: %v", path, err))
	}
}

func rmdir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		panic(fmt.Sprintf("Unable to remove directory: %s: %v", path, err))
	}
}

func pretty(duration time.Duration) string {
	HOURS_PER_YEAR := 24 * 30 * 365
	HOURS_PER_MONTH := 24 * 30
	HOURS_PER_DAY := 24
	hoursAgo := int(duration.Hours())
	if hoursAgo > HOURS_PER_YEAR {
		return fmt.Sprintf("%d years ago", (hoursAgo / HOURS_PER_YEAR))
	} else if hoursAgo > HOURS_PER_MONTH {
		return fmt.Sprintf("%d months ago", (hoursAgo / HOURS_PER_MONTH))
	} else if hoursAgo > 2*HOURS_PER_DAY {
		return fmt.Sprintf("%d days ago", (hoursAgo / HOURS_PER_DAY))
	} else if hoursAgo > HOURS_PER_DAY {
		return "yesterday"
	} else {
		return "today"
	}
}

func errMissingField(field string) error {
	return errors.New(fmt.Sprintf("Missing required field: %s\n", field))
}

func errBlockWord(field string, word string) error {
	return errors.New(fmt.Sprintf("Skipping: %s content contains block word: %s", field, word))
}

func readablePost(url string) string {
	fmt.Printf("GET %s for readability\n", url)
	article, err := readability.FromURL(url, 30*time.Second)
	if err != nil {
		return ""
	}
	return article.TextContent
}

func isEmpty(s string) bool {
	return 0 == len(strings.TrimSpace(s))
}

func readable(html string) string {
	fakeURL, err := url.Parse("")
	if err != nil {
		panic("Unable to build fake URL")
	}
	article, err := readability.FromReader(strings.NewReader(html), fakeURL)
	if err != nil {
		return ""
	}
	return article.TextContent
}

func parseConfig() Config {
	content, err := os.ReadFile("feeds.yaml")
	if err != nil {
		panic(fmt.Sprintf("Unable to open file %e", err))
	}
	config := Config{}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		panic(fmt.Sprintf("Config unmarshal error %e", err))
	}
	return config
}

func safeGUID(post Frontmatter) string {
	id := post.Params.Post.Link
	if post.Params.Post.GUID != "" {
		id = post.Params.Post.GUID
	}
	b := md5.Sum([]byte(id))
	s := hex.EncodeToString(b[:])
	return strings.Replace(s, "=", "", -1)
}

func sortAndLimitPosts(posts []Frontmatter, limit int) []Frontmatter {
	sort.Slice(
		posts,
		func(i, j int) bool {
			return posts[i].Date < posts[j].Date
		},
	)
	if limit > len(posts) {
		return posts
	}
	return posts[:limit]
}

func truncateText(s string, max int) string {
	if max > len(s) {
		return s
	}
	return s[:strings.LastIndexAny(s[:max], " .,:;-")]
}

func isDomainOrSubdomain(questionURL string, domain string) bool {
	questionURL = strings.ToLower(questionURL)
	domain = strings.ToLower(domain)
	u, err := url.Parse(questionURL)
	if err != nil {
		return false
	}
	if u.Host == domain {
		return true
	}
	dotDomain := "." + domain
	if strings.HasSuffix(u.Host, dotDomain) {
		return true
	}
	return false
}
