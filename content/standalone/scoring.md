---
title: Scoring Criteria
url: /scoring/index.html
---

The current scoring methodology aims to promote adoption of features that will support the social aspects of
the blogroll network.
A higher score doesn't mean the blog is more interesting to read, or has quality content,
just that it has adopted features that support the blogroll network.
Future recommendation system-like features are planned.


## Promotes others

***Up to 10 points***

A blog should have an OPML blogroll that promotes other blogs which author enjoys reading.
Peer recommendations are the core of the blogroll network.
Each recommendation earns a point.

See [the blog post](https://alexsci.com/blog/blogroll-network/#deploy-your-own-discoverable-blogroll) for more.


## Promoted by others

***Up to 5 points***

Being promoted by others is a strong quality signal and ensures your blog is discoverable from the network.
Peer recommendations are the core of the blogroll network.
Each recommendation earns one point.

You'll need other sites to adopt OPML blogrolls and promote your site to earn these points.
Ask your friends.


## Has a linked website

***Up to 2 points***

A blog should link to it's feed using the `<link rel="alternate" ...>` syntax (1 point).
The feed should backlink to the website (1 point).
These links help readers identify feeds and helps crawlers associate feeds with their websites.


For example, to link your RSS feed from your website you'd include this code in the head of your pages:

    <link rel="alternate" type="application/rss+xml" href="https://example.com/feed.xml" title="RSS Feed">

Then in your RSS feed you'd link back to the website:

    <link>https://example.com</link>


If you are using a blogging framework, these should be automatically handled.


## Has rel=me links

***Up to 2 points***

A website should link to other related websites using the `<link rel="me" ...>` syntax.
This helps readers find your content across any websites or social media platforms you use.
A `rel=me` link earns a point and a backlink (which verifies the link) earns another point.

[Learn about rel=me](https://microformats.org/wiki/rel-me).


## Has feed categories

***Up to 5 points***

A feed should include categories to help readers understand the themes of the blog.
Each category earns one point.

[Learn about post categories](https://www.rssboard.org/rss-specification#ltcategorygtSubelementOfLtitemgt)
and read the [blog post](https://alexsci.com/blog/rss-categories/) about how categories are used.

## Has post categories

***Up to 3 points***

A post should include categories to help readers understand the themes of each post.
Each category tag on the latest post earns one point.

[Learn about post categories](https://www.rssboard.org/rss-specification#ltcategorygtSubelementOfLtitemgt)
and read the [blog post](https://alexsci.com/blog/rss-categories/) about how categories are used.


## Has a feed title

***3 points***

A feed should have a title to help readers identify the blog.


## Has a feed description

***3 points***

A feed should have a description to help readers understand what the blog is about.

