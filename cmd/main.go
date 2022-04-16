package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/patrickbucher/checklinks"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: checklinks [url]")
		os.Exit(1)
	}
	pageAddr := args[0]
	if !strings.HasPrefix(pageAddr, "http://") && !strings.HasPrefix(pageAddr, "https://") {
		pageAddr = "http://" + pageAddr
	}
	pageURL, err := url.Parse(pageAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse %s as URL: %v", pageAddr, err)
		os.Exit(1)
	}
	crawlPage(pageURL, true, true, true)
}

type Link struct {
	URL  *url.URL
	Site *url.URL
}

func NewLink(address string, site *url.URL) (*Link, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	return &Link{URL: u, Site: site}, nil
}

func (l *Link) IsInternal() bool {
	return l.URL.Hostname() == l.Site.Hostname() || l.URL.Hostname() == ""
}

func (l *Link) IsCrawlable() bool {
	return l.URL.Scheme == "https" || l.URL.Scheme == "http" || l.URL.Scheme == ""
}

type Result struct {
	Err  error
	Link *Link
}

func (c Result) String() string {
	if c.Err != nil {
		return fmt.Sprintf("FAIL %s: %v", c.Link.URL.String(), c.Err)
	} else {
		return fmt.Sprintf("OK %s", c.Link.URL.String())
	}
}

type Sink chan<- struct{}

func crawlPage(site *url.URL, reportOK, reportIgnored, reportFailed bool) {
	var wg sync.WaitGroup
	links := make(chan *Link)
	results := make(chan *Result)
	done := make(chan struct{})

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	go func() {
		visited := make(map[string]struct{})
		for {
			select {
			case l := <-links:
				u := l.URL.String()
				visited[u] = struct{}{}
				if l.IsInternal() {
					l.URL = checklinks.QualifyInternalURL(site, l.URL)
					wg.Add(1)
					go extractLinks(client, l, links, results, done)
				} else {
					wg.Add(1)
					go checkHead(client, l, results, done)
				}
			case result := <-results:
				if result.Err != nil && reportFailed {
					fmt.Println(result)
				}
				if result.Err == nil && reportOK {
					fmt.Println(result)
				}
			case <-done:
				wg.Done()
			}
		}
	}()

	links <- &Link{site, site}
	wg.Wait()
}

func extractLinks(c *http.Client, site *Link, links chan<- *Link, results chan<- *Result, done Sink) {
	u := site.URL.String()
	doc, err := checklinks.FetchDocument(u, c)
	if err != nil {
		results <- &Result{Err: err, Link: site}
		done <- struct{}{}
		return
	}
	hrefs := checklinks.ExtractTagAttribute(doc, "a", "href")
	for _, href := range hrefs {
		link, err := NewLink(href, site.Site)
		if err != nil {
			results <- &Result{Err: err, Link: site}
			continue
		}
		if !link.IsCrawlable() {
			results <- &Result{Err: errors.New("not crawlable"), Link: site}
			continue
		}
		links <- link
	}
	results <- &Result{Err: nil, Link: site}
	done <- struct{}{}
}

func checkHead(c *http.Client, link *Link, results chan<- *Result, done chan<- struct{}) {
	u := link.URL.String()
	response, err := c.Head(u)
	if err != nil {
		results <- &Result{fmt.Errorf("HEAD %v %s", err, u), link}
	} else if response.StatusCode != http.StatusOK {
		statusCode := response.StatusCode
		statusText := http.StatusText(statusCode)
		results <- &Result{fmt.Errorf("HEAD %d %s %s", statusCode, statusText, u), link}
	} else {
		results <- &Result{nil, link}
	}
	done <- struct{}{}
}
