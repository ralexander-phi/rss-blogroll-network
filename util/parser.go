package main

import (
	"bytes"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/atom"
	"github.com/mmcdole/gofeed/json"
	"github.com/mmcdole/gofeed/rss"
	"io"
	"net/http"
)

type CustomParser struct {
	*gofeed.Parser
	ap             *atom.Parser
	jp             *json.Parser
	rp             *rss.Parser
	userAgent      string
	client         *http.Client
	atomTranslator *gofeed.DefaultAtomTranslator
	jsonTranslator *gofeed.DefaultJSONTranslator
	rssTranslator  *gofeed.DefaultRSSTranslator
}

func identifyFeed(feed io.Reader) (gofeed.FeedType, io.Reader) {
	// Wrap the feed io.Reader in a io.TeeReader
	// so we can capture all the bytes read by the
	// DetectFeedType function and construct a new
	// reader with those bytes intact for when we
	// attempt to parse the feeds.
	var buf bytes.Buffer
	tee := io.TeeReader(feed, &buf)
	feedType := gofeed.DetectFeedType(tee)

	// Glue the read bytes from the detect function
	// back into a new reader
	r := io.MultiReader(&buf, feed)
	return feedType, r
}

func mergeFeeds(feed *gofeed.Feed, rss *rss.Feed) []MultiTypeItem {
	out := []MultiTypeItem{}
	for _, item := range feed.Items {
		converted := MultiTypeItem{}
		converted.feed = feed
		converted.item = item
		if rss != nil {
			for _, rssItem := range rss.Items {
				fmt.Printf("Checking: %v \n %v\n", rssItem.GUID, rssItem)
				if rssItem.GUID != nil {
					if item.GUID == rssItem.GUID.Value {
						converted.rss = rssItem
					}
				} else if item.Link == rssItem.Link {
					converted.rss = rssItem
				}
			}
		}
		out = append(out, converted)
	}
	return out
}

func (f *CustomParser) Parse(feed io.Reader) ([]MultiTypeItem, error) {
	var err error
	var rssFeed *rss.Feed
	var parsedFeed *gofeed.Feed

	feedType, r := identifyFeed(feed)

	switch feedType {
	case gofeed.FeedTypeAtom:
		parsed, err := f.ap.Parse(r)
		if err != nil {
			return nil, err
		}
		parsedFeed, err = f.atomTranslator.Translate(parsed)

	case gofeed.FeedTypeJSON:
		parsed, err := f.jp.Parse(r)
		if err != nil {
			return nil, err
		}
		parsedFeed, err = f.jsonTranslator.Translate(parsed)

	case gofeed.FeedTypeRSS:
		rssFeed, err = f.rp.Parse(r)
		if err != nil {
			return nil, err
		}
		parsedFeed, err = f.rssTranslator.Translate(rssFeed)

	default:
		return nil, gofeed.ErrFeedTypeNotDetected
	}

	return mergeFeeds(parsedFeed, rssFeed), nil
}

func (f *CustomParser) ParseURLExtended(url string) ([]MultiTypeItem, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp != nil {
		defer func() {
			closeErr := resp.Body.Close()
			if closeErr != nil {
				err = closeErr
			}
		}()
	}

	return f.Parse(resp.Body)
}

func NewParser() *CustomParser {
	return &CustomParser{
		userAgent:      "Feed2Pages/1.0",
		client:         &http.Client{},
		ap:             &atom.Parser{},
		atomTranslator: &gofeed.DefaultAtomTranslator{},
		jp:             &json.Parser{},
		jsonTranslator: &gofeed.DefaultJSONTranslator{},
		rp:             &rss.Parser{},
		rssTranslator:  &gofeed.DefaultRSSTranslator{},
	}
}
