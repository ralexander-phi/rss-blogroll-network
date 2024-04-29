package main

import (
	"bytes"
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/atom"
	"github.com/mmcdole/gofeed/json"
	"github.com/mmcdole/gofeed/rss"
	"io"
)

type CustomParser struct {
	*gofeed.Parser
	ap             *atom.Parser
	jp             *json.Parser
	rp             *rss.Parser
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
		converted.item = item
		if rss != nil {
			for _, rssItem := range rss.Items {
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

func (f *CustomParser) Parse(feed io.Reader) (*gofeed.Feed, []MultiTypeItem, error) {
	var err error
	var rssFeed *rss.Feed = nil
	var parsedFeed *gofeed.Feed

	feedType, r := identifyFeed(feed)

	switch feedType {
	case gofeed.FeedTypeAtom:
		parsed, err := f.ap.Parse(r)
		if err != nil {
			return nil, nil, err
		}
		parsedFeed, err = f.atomTranslator.Translate(parsed)

	case gofeed.FeedTypeJSON:
		parsed, err := f.jp.Parse(r)
		if err != nil {
			return nil, nil, err
		}
		parsedFeed, err = f.jsonTranslator.Translate(parsed)

	case gofeed.FeedTypeRSS:
		rssFeed, err = f.rp.Parse(r)
		if err != nil {
			return nil, nil, err
		}
		parsedFeed, err = f.rssTranslator.Translate(rssFeed)

	default:
		return nil, nil, gofeed.ErrFeedTypeNotDetected
	}

	posts := mergeFeeds(parsedFeed, rssFeed)
	return parsedFeed, posts, nil
}

func (f *CustomParser) ParseURLExtended(url string) (*gofeed.Feed, []MultiTypeItem, error) {
	resp, err := httpGet(url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Close()
	return f.Parse(resp)
}

func NewParser() *CustomParser {
	return &CustomParser{
		ap:             &atom.Parser{},
		atomTranslator: &gofeed.DefaultAtomTranslator{},
		jp:             &json.Parser{},
		jsonTranslator: &gofeed.DefaultJSONTranslator{},
		rp:             &rss.Parser{},
		rssTranslator:  &gofeed.DefaultRSSTranslator{},
	}
}
