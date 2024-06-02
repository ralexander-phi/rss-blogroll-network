---
title: Feed Scoring System
---

# Feed Scoring

Each feed is examined by the following criteria to assess a feed score.
The feed owner can improve their feed score by making adjustments to their site.

Criteria are intended to improve reader experience and discoverability of the feed.

Over time new criteria will be added and the weights may be adjusted.

### OPML Blogroll

**Up to 10 points**

Adding an [OPML Blogroll](https://opml.org/blogroll.opml) is a great way to promote other websites and blogs you follow.
The OPML format standardizes blogrolls so that RSS software can discover your recommendations.

To gain these points, add a blogroll and begin promoting websites.
Each website you promote adds half a point, up to the maximum of 10 points.

Promoting feeds in your OPML Blogroll helps us grow the blogroll network.


### Title and description

**Up to 6 points**

Make sure you give your feed a title and description.

### Feed categories

**Up to 5 points**

Help readers understand what kind of content you have in your feed by adding categories.
You'll get a point for each category you add, up to five points.


### Reverse OPML Blogroll

**Up to 5 points**

You'll need another website owner to promote your website in their OPML blogroll.
Getting recommended by another site in the network will earn you 5 points.


### Post categories

**Up to 3 points**

We share your latest post's title and description to help readers understand what your blog is about.
Tag your posts with categories to earn up to 3 points.


### Link to your website

**Up to 2 points**

Your web feed should point to your website and your website should link to your RSS feed.
You need both links present to get both points.


### Related webpages

**Up to 2 points**

Help readers find you elsewhere on the web by linking other social profiles (Mastodon, Twitter, GitHub, etc).

This website discovers other websites by detecting `<link rel="me" ...>` or `<a rel="me" ...>` elements on your web page.
You'll get a point for listing at least one other site.

If the linked page has a `rel=me` backlink to your site, then the link is verified.
You get a total of two points if you have at least one verified related webpage.


## Debugging

This website and the scoring system are still being developed, so you may see some issues.

Here's some things to consider if your feed page doesn't look right:

* The crawler is a bot, so it can't solve captchas
* I usually crawl from VPN or cloud IP addresses, ensure you aren't blocking these
* The crawler respects `robots.txt` settings
* The crawler ignores pages that set `noindex`
* The crawler runs infrequently (I'll try to run it at least weekly)
  * Check the "last refreshed" timestamp in the page footer

If you're still not sure, [open an issue](https://github.com/ralexander-phi/rss-blogroll-network/issues).


## Non-criteria

Here's some things we won't make you do for points:

* Backlink to this site - although this is always appreciated
* Have fresh content - just post when you have something to say
