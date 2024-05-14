# RSS Blogroll Networks

RSS feeds can programmatically define a [blogroll using OPML](https://opml.org/blogroll.opml).
These blogrolls help people who read your blog discover other websites you think are worth promoting.

This project maps connections between blogs and other RSS-enabled entities.
The size and interconnectedness of the network can be tracked over time.

[Read the blog post.](https://alexsci.com/blog/blogroll-network/)

## Joining the network

The best way to join the network is for a blog already in the network to add your blog to their blogroll.
This may happen organically over time if you have content that's interesting to others in the network.

To help the network get bootstrapped, and discover disjoint parts of the network, you can open a GitHub Issue to add you blog.
You'll qualify if:

* You have an RSS feed
* You have an OPML blogroll
  * It promotes at least two blogs or websites
  * Your OPML blogroll is discoverable:
    * As an element of your RSS feed; or
    * As a link on your website
* Your site is personal, non-commercial, and ad-free
* Your site has human generated content
* Content may be in any language
* Content is reasonably "safe for work"
* You aren't blocking us via `robots.txt`


## Opt out of the network

We'll respect your decision if you don't want your website listed here.

Our crawler uses the `Feed2Pages/*` User-Agent string and respects `robots.txt`.
Block this User-Agent (or all bots) from accessing your RSS feed using your `robots.txt` file.
