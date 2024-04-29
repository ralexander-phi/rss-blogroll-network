package main

import (
	"bufio"
	"errors"
	"fmt"
	readability "github.com/go-shiori/go-readability"
	"github.com/go-yaml/yaml"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	return resp.Body, nil
}

func readUrl(url string) (io.Reader, io.Closer, error) {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		f, err := httpGet(url)
		if err != nil {
			return nil, nil, err
		}
		return bufio.NewReader(f), f, nil
	} else if strings.HasPrefix(url, "file://") {
		url := strings.Replace(url, "file://", "", 1)
		return readFile(url)
	}
	return nil, nil, errors.New(fmt.Sprintf("Unsupported protocol: %s", url))
}

func readFile(path string) (io.Reader, io.Closer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewReader(f), f, nil
}

func writeYaml(o any, path string) {
	output, err := yaml.Marshal(o)
	if err != nil {
		panicStringErr("YAML error", err)
	}

	// Markdown uses `---` for YAML frontmatter
	sep := []byte("---\n")
	output = append(sep, output...)
	output = append(output, sep...)

	err = os.WriteFile(path, output, os.FileMode(int(0660)))
	if err != nil {
		panicStringsErr("Unable to write file", path, err)
	}
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
