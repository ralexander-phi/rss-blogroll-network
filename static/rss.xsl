<?xml version="1.0" encoding="utf-8"?>
<xsl:stylesheet version="3.0"
	xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
	xmlns:atom="http://www.w3.org/2005/Atom">
	<xsl:output method="html" version="1.0" encoding="UTF-8" indent="yes"/>
	<xsl:template match="/">
	<xsl:variable name="owner_url"><xsl:value-of select="/rss/channel/link"/></xsl:variable>

  <!-- Kudos to: https://github.com/fileformat/rss.style (MIT License) -->

<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<meta name="referrer" content="unsafe-url" />
		<title><xsl:value-of select="/rss/channel/title"/></title>
		<link rel="stylesheet" href="https://www.rss.style/css/water.min.css" />
	</head>
	<body>
		<h1>
			<img alt="feed icon" src="https://www.vectorlogo.zone/logos/rss/rss-tile.svg" style="height:1em;vertical-align:middle;" />&#xa0;
			<xsl:value-of select="/rss/channel/title"/>
		</h1>

		<p>
			<xsl:value-of select="/rss/channel/description"/>
		</p>

		<p>This is the RSS&#xa0;<a href="https://www.rss.style/what-is-a-feed.html">news feed</a>&#xa0;for the&#xa0;
			<a><xsl:attribute name="href">
				<xsl:value-of select="/rss/channel/link"/>
			</xsl:attribute>
			<xsl:value-of select="/rss/channel/title"/></a>&#xa0;
			website.
		</p>

		<p>It is meant for&#xa0;<a href="https://www.rss.style/newsreaders.html">news readers</a>, not humans.  Please copy-and-paste the URL into your news reader!</p>

		<p>
			<pre>
				<code id="feedurl"><xsl:value-of select="/rss/channel/atom:link/@href"/></code>
			</pre>
			<button
				class="clipboard"
				data-clipboard-target="#feedurl">
				Copy to clipboard
			</button>
		</p>

		<xsl:for-each select="/rss/channel/item">
			<details open="true"><summary>
				<a>
					<xsl:attribute name="href">
						<xsl:value-of select="link"/>
					</xsl:attribute>
					<xsl:value-of select="title"/>
				</a>&#xa0;-&#xa0;
				<xsl:value-of select="pubDate" />
				</summary>
				<xsl:value-of select="description" disable-output-escaping="yes" />
				</details>
		</xsl:for-each>
		<p><xsl:value-of select="count(/rss/channel/item)"/> news items.</p>
		<p><small>Powered by <a href="https://www.rss.style/"><img referrerpolicy="origin" src="https://www.rss.style/favicon.svg" style="height:1em;padding-right:0.25em;vertical-align:middle;" />RSS.Style</a></small></p>
		<script src="https://cdn.jsdelivr.net/npm/clipboard@2.0.11/dist/clipboard.min.js"></script>
		<script>
			new ClipboardJS('.clipboard');
		</script>
	</body>
</html>
	</xsl:template>
</xsl:stylesheet>
