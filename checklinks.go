package checklinks

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

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
