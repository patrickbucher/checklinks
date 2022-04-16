package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

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
	crawlPage(pageURL)
}

type rChan (<-chan *url.URL)
type wChan (chan<- *url.URL)
type sChan (chan<- string)

func crawlPage(site *url.URL) {
	var wg sync.WaitGroup
	worklist := make(chan *url.URL)
	internal := make(chan *url.URL)
	external := make(chan *url.URL)
	ignored := make(chan string)
	done := make(chan struct{})

	push := func(u *url.URL) {
		worklist <- u
	}

	go func(wInt rChan, wExt rChan, ign <-chan string, d <-chan struct{}) {
		for {
			select {
			case u := <-wInt:
				wg.Add(1)
				go push(u)
			case u := <-wExt:
				response, err := http.Head(u.String())
				if err != nil {
					fmt.Fprintf(os.Stderr, "HEAD %s: %v\n", u.String(), err)
				} else {
					fmt.Println("HEAD", response.StatusCode, u.String())
				}
			case s := <-ign:
				fmt.Println("IGNORE", s)
			case <-d:
				return
			}
		}
	}(internal, external, ignored, done)

	f := extractLinks(site, &wg)
	go f(worklist, internal, external, ignored)

	wg.Add(1)
	go push(site)

	wg.Wait()
	done <- struct{}{}
	close(worklist)
	close(internal)
	close(external)
	close(ignored)
	close(done)
}

func extractLinks(site *url.URL, wg *sync.WaitGroup) func(rChan, wChan, wChan, sChan) {
	visited := make(map[string]struct{})
	return func(in rChan, outInt, outExt wChan, outIgnore sChan) {
		for u := range in {
			fmt.Println("TODO", u)
			if _, ok := visited[u.String()]; ok {
				wg.Done()
				continue
			}
			doc, err := checklinks.FetchDocument(u.String())
			fmt.Println("GET", 200, u.String())
			if err != nil {
				log.Println(err)
				outIgnore <- u.String() // TODO: proper error structure with more information
				wg.Done()
				continue
			}
			hrefs := checklinks.ExtractTagAttribute(doc, "a", "href")
			for _, href := range hrefs {
				u, err := url.Parse(href)
				if err != nil {
					outIgnore <- href
					continue
				}
				if u.Scheme != "https" && u.Scheme != "http" && u.Scheme != "" {
					outIgnore <- href
					continue
				}
				if u.Hostname() == site.Hostname() || u.Hostname() == "" {
					outInt <- checklinks.QualifyInternalURL(site, u)
				} else {
					outExt <- u
				}
			}
			visited[u.String()] = struct{}{}
			wg.Done()
		}
	}
}
