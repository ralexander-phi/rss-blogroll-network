---
title: Privacy
url: /privacy/index.html
---

Web scraping and indexing is an inherently privacy invasive practice.
As such, this project uses several methods to reduce unwanted impacts while achieving the goal
of improving discoverability of personal blogs, and their feeds.

The privacy impacts of this project and controls are described here.

## Terms

The Website - this website at https://alexsci.com/rss-blogroll-network/.

The Crawler - the web crawler that collects data displayed on The Website.

The Project - The Website, The Crawler, and associated works.


## Data collected by The Website

The Website collects metrics about visitors using
[GoatCounter](https://www.goatcounter.com) (a third-party).
GoatCounter was selected as it is open source, collects metrics in a privacy preserving way, and does not use any cookies.
The [GoatCounter docs](https://www.goatcounter.com) explain that GDPR consent is likely not needed as it doesn't collect any
personally identifying information.

Here's the settings being used on GoatCounter:

{{< figure
    src="/GoatCounterSettings.png"
    title="From the GoatCounter settings page" 
    alt="Checked items: sessions, referrer, user-agent, size, country. Unchecked: region, language."
>}}

This collection is helpful in understanding the technology used to access The Website and
the popularity of various pages.

You can opt-out of this data collection by installing a web browser extension like
[uBlock Origin](https://github.com/gorhill/uBlock#ublock-origin).
The default uBO configuration blocks GoatCounter (and many other types of content).
Extensions like uBO work best on [Firefox](https://www.mozilla.org/en-US/firefox/new/).


## Hosting

The Website is hosted on GitHub Pages, which has it's own [privacy policy](https://docs.github.com/en/pages/getting-started-with-github-pages/about-github-pages#data-collection).


## Cookies, tracking pixels, local storage, CDNs 

These are not used on The Website.


## Data collected by The Crawler

The Crawler may collect the following data, as provided by a public available web page.
The following types of information are collected:

* Title
* Descriptions
* Category
* Links

Body content of web pages, like the text of a blog post, is not collected.

The Crawler will collect Personally Identifiable Information if it is present in a collected field.
For example, my blog is named "Robert Alexander's Blog", which has my name in the title.


## Opt-outs for The Crawler

You may use the following methods to opt-out your web content out of being indexed by
The Project.
These are industry standards and are useful to implement if you'd like to control how
standards compliant web crawlers interact with your website.


### Manual request

You can always manually request for your site to be excluded from the project.
Your domain will be listed under [`block_domains`](https://github.com/ralexander-phi/rss-blogroll-network/blob/main/feeds.yaml).
Content related to your web pages will be removed after the crawlers next run.

Contact methods:

* [Open a GitHub Issue](https://github.com/ralexander-phi/rss-blogroll-network)
* [DM me on Mastodon](https://indieweb.social/@robalex)
* Send me an email: robert at robalexdev dot com

You will need to demonstrate that you are the owner the requested domain.
This is a personal project, so I'll process any requests as I am available.
No timeline is provided, although I consider opt-outs as priority incidents.


### robots.txt

A `robots.txt` file hosted at the root of your domain (I.E. https://example.com/robots.txt) can
be used to control what content various automated user agents are allowed to access.

For example, if you don't want any web crawlers to access your site, you can block them all using:

    User-agent: *
    Disallow: /

The Crawler (and any other well-behaved web crawler) will not access any content on
your site (except the robots.txt file) when you use this setting.

If you'd like to block every crawler other than The Crawler, you can write:

    User-agent: *
    Disallow: /
    User-agent: Feed2Pages/0.1
    Disallow:

If you'd like to only block The Crawler use:

    ...
    User-agent: Feed2Pages/0.1
    Disallow: /

More fine-grained control of access is also possible.
For example, if you'd like The Crawler to process your Atom feed but not your RSS feed you can
write something like:

    ...
    User-agent: Feed2Pages/0.1
    Disallow: /rss.xml


### noindex tag

The `noindex` tag instructs web crawlers not to index a page.
You can ask all crawlers not to index your page by placing the following HTML inside your `<head>` section:

    <meta name="robots" content="noindex">

You can selectively ask The Project not to index your page using:

    <meta name="feed2pages/0.1" content="noindex">

Or you can indicate that you only want certain crawlers to index your page:

    <meta name="robots" content="noindex">
    <meta name="feed2pages/0.1" content="">

You can also put `noindex` in an HTTP header:

    X-Robots-Tag: noindex

Any page with `noindex` set will not be shown on The Website.
As crawling is an intermittent process, pages may remain on The Website until after the next crawl is completed.

You can read more about [`noindex` in Google's documentation](https://developers.google.com/search/docs/crawling-indexing/block-indexing).


### noarchive tag

While The Project doesn't operate an archive,
[other sites](https://github.com/ralexander-phi/rss-blogroll-network/issues/8)
may use The Project's data as part of their archival process.

When The Crawler sees the `noarchive` tag on a feed, the `noarchive` tag will be used on any pages on
The Website that were generated for that feed.
Note that RSS and Atom feeds are not HTML documents, so you'll need to use the HTTP header approach mentioned above.
Archivers like the [Internet Archive](https://archive.org/post/31561/robots-archive-noarchive-meta-tags)
respect the `noarchive` tag.


### HTTP status codes

The Crawler will not index content that is restricted.
A HTTP request that returns a [401 Unauthorized](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
status code, for example, will not be indexed.

You can selectively block The Crawler by detecting the User Agent String
([`Feed2Pages/0.1`](https://github.com/ralexander-phi/feed2pages-action/blob/main/const.go#L3))
and returning
[one of the supported status codes](https://github.com/ralexander-phi/feed2pages-action/blob/ffc36d4ff5827d3f8db2ad2f7ec9a47dc30ff2a3/crawler.go#L27).


## Use of collected data

The Website displays collected data for ease of browsing.
The public data collected by the crawler is an [open data set and is publicly available](https://alexsci.com/rss-blogroll-network/index.json) on The Website.
Expected additional uses include recommendation and discovery systems for RSS readers.
As an open data set, others may use the data in other ways.


## Open source code

The crawler is open source and the code is available for review:

* [The Crawler](https://github.com/ralexander-phi/feed2pages-action)
* [The Website](https://github.com/ralexander-phi/rss-blogroll-network)

You may inspect the behavior in detail.


## Errata and changes

This page will be updated as privacy impacting changes occur.
This project is run by a human, I occasionally make mistakes, let me know if you see any
bugs, errors or omissions.

