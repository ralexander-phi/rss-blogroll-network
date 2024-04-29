package main

import (
	"encoding/xml"
	"fmt"
)

type XmlOpml struct {
	Body XmlOpmlBody `xml:"body"`
	Head XmlOpmlHead `xml:"head"`
}

type XmlOpmlHead struct {
	Title string `xml:"title"`
}

type XmlOpmlBody struct {
	Outline []XmlOutline `xml:"outline"`
}

type XmlOutline struct {
	XMLName xml.Name     `xml:"outline"`
	Text    string       `xml:"text,attr"`
	Type    string       `xml"type,attr"`
	XmlUrl  string       `xml:"xmlUrl,attr"`
	Title   string       `xml:"title,attr"`
	Outline []XmlOutline `xml:"outline"`
}

func asDetails(outlines []XmlOutline) []FeedDetails {
	result := []FeedDetails{}
	for _, outline := range outlines {
		if outline.XmlUrl != "" {
			details := FeedDetails{}
			details.Link = outline.XmlUrl
			details.Text = outline.Text
			details.Type = outline.Type
			details.Text = outline.Text
			result = append(result, details)
		}
		result = append(result, asDetails(outline.Outline)...)
	}
	return result
}

func parseOpml(url string) []FeedDetails {
	content, closer, err := readUrl(url)
	if err != nil {
		panicStringsErr("Error reading URL", url, err)
	}
	defer closer.Close()
	decoder := xml.NewDecoder(content)
	opml := XmlOpml{}
	err = decoder.Decode(&opml)
	if err != nil {
		panicStringErr("Error decoding OMPL", err)
	}
	fmt.Printf("OPML parsed as: %v\n", opml)
	return asDetails(opml.Body.Outline)
}
