package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-yaml/yaml"
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
	err := os.MkdirAll(path, 0770)
	if err != nil {
		panicStringsErr("Unable to create directory", path, err)
	}
}

func rmdir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		panicStringsErr("Unable to remove directory", path, err)
	}
}

func pluralizeAgo(s string, i int) string {
	if i == 1 {
		return fmt.Sprintf("one %s ago", s)
	} else {
		return fmt.Sprintf("%d %ss ago", i, s)
	}
}

func pretty(duration time.Duration) string {
	HOURS_PER_DAY := 24
	HOURS_PER_MONTH := HOURS_PER_DAY * 30
	HOURS_PER_YEAR := HOURS_PER_MONTH * 12
	hoursAgo := int(duration.Hours())
	if hoursAgo > HOURS_PER_YEAR {
		t := hoursAgo / HOURS_PER_YEAR
		return pluralizeAgo("year", t)
	} else if hoursAgo > HOURS_PER_MONTH {
		t := hoursAgo / HOURS_PER_MONTH
		return pluralizeAgo("month", t)
	} else if hoursAgo > 2*HOURS_PER_DAY {
		t := hoursAgo / HOURS_PER_DAY
		return pluralizeAgo("day", t)
	} else if hoursAgo > HOURS_PER_DAY {
		return "yesterday"
	} else {
		return "today"
	}
}

func firstNonEmpty(options []string) string {
	for _, option := range options {
		if !isEmpty(option) {
			return strings.TrimSpace(option)
		}
	}
	return ""
}

func isEmpty(s string) bool {
	return 0 == len(strings.TrimSpace(s))
}

func parseConfig() Config {
	content, closer, err := readFile("feeds.yaml")
	if err != nil {
		panicStringErr("Unable to parse config", err)
	}
	defer closer.Close()
	config := Config{}
	decoder := yaml.NewDecoder(content)
	err = decoder.Decode(&config)
	if err != nil {
		panicStringErr("Config decode error", err)
	}
	// Parse the OPML file (local file or remote resource)
	config.Feeds = parseOpml(config.FeedUrl)
	return config
}

func safeGUID(post PostFrontmatter) string {
	id := post.Params.Post.Link
	if post.Params.Post.GUID != "" {
		id = post.Params.Post.GUID
	}
	b := md5.Sum([]byte(id))
	s := hex.EncodeToString(b[:])
	return strings.Replace(s, "=", "", -1)
}

func sortAndLimitPosts(posts []PostFrontmatter, limit int) []PostFrontmatter {
	sort.Slice(
		posts,
		func(i, j int) bool {
			return posts[i].Date > posts[j].Date
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

func panicStringErr(s string, err error) {
	panic(fmt.Sprintf("%s: %e", s, err))
}

func panicStringsErr(s1, s2 string, err error) {
	panic(fmt.Sprintf("%s: %s: %e", s1, s2, err))
}

func panicErr(err error) {
	panic(fmt.Sprintf("%e", err))
}

func errMissingField(field string) error {
	return errors.New(fmt.Sprintf("Missing required field: %s\n", field))
}

func errBlockWord(field string, word string) error {
	return errors.New(fmt.Sprintf("Skipping: %s content contains block word: %s", field, word))
}

func unixEpoc() time.Time {
	return time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
}
