package main

import (
	"fmt"
	readability "github.com/go-shiori/go-readability"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func httpGet(url string) (io.ReadCloser, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Feed2Pages/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func readUrlOrPanic(url string) []byte {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		body, err := httpGet(url)
		if err != nil {
			panicErr(err)
		}
		defer body.Close()
		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			panicErr(err)
		}
		return bodyBytes
	} else if strings.HasPrefix(url, "file://") {
		url := strings.Replace(url, "file://", "", 1)
		return readFileOrPanic(url)
	}
	panic(fmt.Sprintf("Unsupported protocol: %s", url))
}

func readFileOrPanic(path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		panicStringsErr("Unable to open file", path, err)
	}
	return content
}

func readablePost(url string) string {
	fmt.Printf("GET %s for readability\n", url)
	article, err := readability.FromURL(url, 30*time.Second)
	if err != nil {
		return ""
	}
	return article.TextContent
}

func readable(html string) string {
	fakeURL, err := url.Parse("")
	if err != nil {
		panic("Unable to build fake URL")
	}
	article, err := readability.FromReader(strings.NewReader(html), fakeURL)
	if err != nil {
		return ""
	}
	return article.TextContent
}

func isDomainOrSubdomain(questionURL string, domain string) bool {
	questionURL = strings.ToLower(questionURL)
	domain = strings.ToLower(domain)
	u, err := url.Parse(questionURL)
	if err != nil {
		return false
	}
	if u.Host == domain {
		return true
	}
	dotDomain := "." + domain
	if strings.HasSuffix(u.Host, dotDomain) {
		return true
	}
	return false
}
