# Feed2Pages

A blogroll that aggregates RSS feeds into your own news feed web site.


## Use cases

* A personal RSS reader you can access from all your devices
* Share what you're reading and promote the RSS feeds you follow
* Aggregate multiple feeds to create a news feed for a particular topic
* Discover new blogs from the bloggers you follow


## Build your own

See also: [feed2pages-papermod](https://github.com/ralexander-phi/feed2pages-papermod).

First fork this repository into your GitHub account.
Then enable GitHub Pages:

![Steps to enable GitHub Pages](images/Enable-Pages.png)

1. Open the Settings tab on your repo
2. Select Pages from the lefthand navigation
3. Set the source to GitHub Actions

You can configure a custom domain or enforce HTTPS on this page as well.


## Promote your links

Export an OPML from your feed reader.
Upload your OPML export as `https://<your-site>/.well-known/recommendations.opml` or another location.
Finally link to your OPML file using: `<link rel="blogroll" type="text/xml" href="https://<your-site>/.well-known/recommendations.opml">`.
You'll need to re-export this file to pick up any changes to who you follow.

For web-based readers, find a URL for your OPML file and link to that instead of uploading.

Software that supports [this syntax](https://opml.org/blogroll.opml) can help readers of your blog discover even more content.


## Connect your feeds

If you're already using an RSS feed reader, check if it can export an OPML file.
Export this and save it as `site/static/links.opml`.

You can manage your feed on sites like [FeedLand](https://feedland.com), which publishes your subscriptions at `https://feedland.com/opml?screenname=<yourname>`.
Edit `site/feeds.yaml` and set `feed_url` to the URL of your OPML file.

Alternatively, you can edit the provided sample file (`site/static/links.opml`) manually.
The most important field is `xmlUrl` (which points to the feed URL).


## Running locally

`git clone` [feed2pages-action](https://github.com/ralexander-phi/feed2pages-action) and build it

    $ go build

Then in this repo run:

    $ ../feed2pages-action/util
    $ hugo server


Consider editing your `feeds.yaml` to reduce how many feeds you fetch in testing.


## How it works

1. The repository owner configures the RSS feeds they wish to follow in `feeds.yaml`.
2. They configure settings such as block words to curate their news feed
3. GitHub Actions runs as a periodic (daily) cron job:
    1. The scraping utility collects articles from the RSS feeds
    2. The feed contents are normalized and enriched
    3. The discovered feeds and posts are saved as Hugo content
    4. Recommended feeds are discovered iteratively
    5. Hugo builds the site into static HTML
    6. GitHub Actions publishes the HTML to GitHub Pages


## Ideas

* Collect articles from popular news website and aggregator RSS feeds, using filters to create a single topic news feed.
  * Hacker News: https://news.ycombinator.com/rss
  * Hacker News RSS: https://hnrss.org
  * Reddit Subreddits: https://www.reddit.com/r/programming.rss
  * Lobsters: https://lobste.rs/t/programming,compsci.rss
  * NY Times: https://rss.nytimes.com/services/xml/rss/nyt/World.xml
  * Many more: https://github.com/plenaryapp/awesome-rss-feeds/blob/master/README.md
