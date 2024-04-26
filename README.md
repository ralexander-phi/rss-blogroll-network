# GitHub Pages as an RSS reader

[Git scrape](https://simonwillison.net/2020/Oct/9/git-scraping/) RSS feeds into a news feed


## Use cases

* A personal RSS reader you can access from all your devices
* Share what you're reading and the RSS feeds you follow
* Aggregate multiple feeds to create a news feed for a particular topic


## Build your own

First fork this repository into your GitHub account.
Then enable GitHub Pages:

![Steps to enable GitHub Pages](/images/Enable Pages.png).

1. Open the Settings tab on your repo
2. Select Pages from the lefthand navigation
3. Set the source to GitHub Actions

You can configure a custom domain or enforce HTTPS on this page as well.

Edit `site/feeds.yaml` to customize your feed.


## feeds.yaml settings


### Configure feeds

A list of URLs with optional settings:

```
feeds:
  - url: https://example.com/rss.xml
  - url: https://example.co.uk/atom.xml
    block_words:
      - Bitcoin
      - Cryptocurrency
```

The above settings pulls an RSS feed from example.com and an Atom feed from example.co.uk.
Articles that contain block words, "Bitcoin" and "Cryptocurrency", will be filtered out of the example.co.uk feed but not the example.com feed.

Required fields:
* `url`: The URL of the RSS, Atom, or JSON feed

Optional fields:
* `block_words`: Articles that contain any of these words in the title, description, or page content will be filtered out.
* `block_domains`: Articles from this domain, or subdomains of this domain, will be filtered out. (Overrides the global setting)
* `ignore_description`: Some RSS feeds contain metadata in the description field. Setting this to `true` will cause Feed2Pages to generate an alternate description from the RSS `content` field or by scraping the linked article. (Overrides the global setting)


### Global settings

`post_age_limit_days`: Filter out posts older than this limit

`max_posts_per_feed`: Include only the newest N posts from each feed. This helps when some feeds publish content much more frequently than others, as they could otherwise fill the news feed.

`max_posts`: Limit the number of posts to display.

`block_words`: Articles that contain any of these words in the title, description, or page content will be filtered out.

`block_domains`: Articles from this domain, or subdomains of this domain, will be filtered out.

## How it works

1. The repository owner configures the RSS feeds they wish to follow
2. They configure settings such as block words to curate the news feed
3. GitHub Actions runs as a periodic (daily) cron job
4. The scraping utility collects articles from RSS feeds
5. The feed contents are normalized and enriched
6. The discovered articles are saved as [Hugo](https://gohugo.io/) pages
7. Hugo builds the site into static HTML
8. GitHub Actions publishes the HTML to GitHub Pages


## Ideas

* Collect articles from popular news website and aggregator RSS feeds, using filters to create a single topic news feed.
  * Hacker News: https://news.ycombinator.com/rss
  * Hacker News RSS: https://hnrss.org
  * Reddit Subreddits: https://www.reddit.com/r/programming.rss

