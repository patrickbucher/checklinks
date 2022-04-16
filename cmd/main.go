package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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
	doc, err := checklinks.FetchDocument(pageURL.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch document at %s: %v", pageAddr, err)
		os.Exit(1)
	}
	aHrefs := checklinks.ExtractTagAttribute(doc, "a", "href")
	internalHrefs := make([]*url.URL, 0)
	externalHrefs := make([]*url.URL, 0)
	for _, href := range aHrefs {
		u, err := url.Parse(href)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse URL %s: %v\n", href, err)
		}
		if u.Scheme != "https" && u.Scheme != "http" && u.Scheme != "" {
			continue
		}
		if u.Hostname() == pageURL.Hostname() || u.Hostname() == "" {
			// absolute link on same page; or relative link (also on same page)
			fullURL := checklinks.QualifyInternalURL(pageURL, u)
			internalHrefs = append(internalHrefs, fullURL)
		} else {
			externalHrefs = append(externalHrefs, u)
		}
	}
	for _, href := range internalHrefs {
		response, err := http.Get(href.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "GET %s: %v\n", href.String(), err)
			continue
		}
		fmt.Println("GET", response.StatusCode, href.String())
	}
	for _, href := range externalHrefs {
		response, err := http.Head(href.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "HEAD %s: %v\n", href.String(), err)
			continue
		}
		fmt.Println("HEAD", response.StatusCode, href.String())
	}
}
