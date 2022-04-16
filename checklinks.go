package checklinks

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/html"
)

var errNotCrawlable = errors.New("not crawlable")

// FetchDocument gets the document indicated by the given url using the given
// client, and returns its root (document) node. An error is returned if the
// document cannot be fetched or parsed as HTML.
func FetchDocument(url string, c *http.Client) (*html.Node, error) {
	response, err := c.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %v", url, err)
	}
	defer response.Body.Close()
	docNode, err := html.Parse(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parse document at %s: %v", url, err)
	}
	return docNode, nil
}

// ExtractTagAttribute traverses the given node's tree, searches it for nodes
// with the given tag name, and extracts the given attribute value from it.
func ExtractTagAttribute(node *html.Node, tagName, attrName string) []string {
	attributes := make([]string, 0)
	if node.Type != html.ElementNode && node.Type != html.DocumentNode {
		return attributes
	}
	if node.Data == tagName {
		for _, attr := range node.Attr {
			if attr.Key == attrName {
				attributes = append(attributes, attr.Val)
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		childAttributes := ExtractTagAttribute(c, tagName, attrName)
		attributes = append(attributes, childAttributes...)
	}
	return attributes
}

// QualifyInternalURL creates a new URL by merging scheme and host information
// from the page URL with the rest of the URL indication from the link URL.
func QualifyInternalURL(page, link *url.URL) *url.URL {
	qualifiedURL := &url.URL{
		Scheme: page.Scheme,
		Host:   page.Host,
		Path:   link.Path,
		// TODO: Query Parameters?
	}
	return qualifiedURL
}

// Link represents a link (URL) in the context of a web site (Site).
type Link struct {
	URL  *url.URL
	Site *url.URL
}

// NewLink creates a Link from the given address. An error is returned, if the
// address cannot be parsed.
func NewLink(address string, site *url.URL) (*Link, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	return &Link{URL: u, Site: site}, nil
}

// IsInternal returns true if the link's URL points to the same domain as its
// site, and false otherwise.
func (l *Link) IsInternal() bool {
	return l.URL.Hostname() == l.Site.Hostname() || l.URL.Hostname() == ""
}

// IsCrawlable returns true if the URL of the link has http(s) as the protocol,
// or no protocol at all (which indicates an internal link), and false
// otherwise.
func (l *Link) IsCrawlable() bool {
	return l.URL.Scheme == "https" || l.URL.Scheme == "http" || l.URL.Scheme == ""
}

// Result describes the result of processing a Link.
type Result struct {
	Err  error
	Link *Link
}

// String returns a string prefixed with FAIL in case of an error, and prefixed
// with OK if no error is present. The URL and error (if any) is contained in
// the string.
func (c Result) String() string {
	if c.Err != nil {
		return fmt.Sprintf(`FAIL "%s": %v`, c.Link.URL.String(), c.Err)
	} else {
		return fmt.Sprintf(`OK "%s"`, c.Link.URL.String())
	}
}

// Sink represents a writable channel to which empty structs are written.
type Sink chan<- struct{}

// CrawlPage crawls the given site's URL and reports successfully checked
// links, ignored links, and failed links (according to the flags ok, ignore,
// fail, respectively). The given timeout is used to limit the waiting time of
// the http client for a request.
func CrawlPage(site *url.URL, timeout int, ok, ignore, fail bool) {
	var wg sync.WaitGroup
	links := make(chan *Link)
	results := make(chan *Result)
	done := make(chan struct{})

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	go func() {
		visited := make(map[string]struct{})
		for {
			select {
			case l := <-links:
				u := l.URL.String()
				visited[u] = struct{}{}
				if l.IsInternal() {
					l.URL = QualifyInternalURL(site, l.URL)
					wg.Add(1)
					go ProcessNode(client, l, links, results, done)
				} else {
					wg.Add(1)
					go ProcessLeaf(client, l, results, done)
				}
			case result := <-results:
				if result.Err != nil {
					if errors.Is(result.Err, errNotCrawlable) {
						if ignore {
							fmt.Println(result)
						}
					} else if fail {
						fmt.Println(result)
					}
				}
				if result.Err == nil && ok {
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

// ProcessNode uses the given http.Client to fetch the given site, and reports
// the extracted links on the page (indicated by <a href="...">). Links
// unsuitable for further crawling and malformed links are reported. A message
// is sent to the given done channel when the node has been processed.
func ProcessNode(c *http.Client, site *Link, links chan<- *Link, results chan<- *Result, done Sink) {
	u := site.URL.String()
	doc, err := FetchDocument(u, c)
	if err != nil {
		results <- &Result{Err: err, Link: site}
		done <- struct{}{}
		return
	}
	hrefs := ExtractTagAttribute(doc, "a", "href")
	for _, href := range hrefs {
		link, err := NewLink(href, site.Site)
		if err != nil {
			results <- &Result{Err: err, Link: site}
			continue
		}
		if !link.IsCrawlable() {
			results <- &Result{Err: errNotCrawlable, Link: site}
			continue
		}
		links <- link
	}
	results <- &Result{Err: nil, Link: site}
	done <- struct{}{}
}

// ProcessLeaf uses the given http.Client to fetch the given link using a HEAD
// request, and reports the result of that request. If HEAD is not supported,
// GET is tried in addition. A message is sent to the given done channel when
// the node has been processed.
func ProcessLeaf(c *http.Client, link *Link, results chan<- *Result, done chan<- struct{}) {
	u := link.URL.String()
	response, err := headOrGet(c, link.URL)
	if err != nil {
		results <- &Result{Err: err, Link: link}
	} else if response.StatusCode != http.StatusOK {
		statusCode := response.StatusCode
		statusText := http.StatusText(statusCode)
		results <- &Result{fmt.Errorf("HEAD %d %s %s", statusCode, statusText, u), link}
	} else {
		results <- &Result{nil, link}
	}
	done <- struct{}{}
}

func headOrGet(c *http.Client, u *url.URL) (*http.Response, error) {
	response, err := c.Head(u.String())
	if err != nil {
		return nil, fmt.Errorf("HEAD %v %s", err, u.String())
	}
	if response.StatusCode == http.StatusMethodNotAllowed {
		response, err = c.Get(u.String())
		if err != nil {
			return nil, fmt.Errorf("GET %v %s", err, u.String())
		}
	}
	return response, nil
}
