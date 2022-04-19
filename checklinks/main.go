package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/patrickbucher/checklinks"
)

var (
	timeout       = flag.Int("timeout", 10, "request timeout (in seconds)")
	showSucceeded = flag.Bool("success", false, "report succeeded links (OK)")
	showIgnored   = flag.Bool("ignored", false, "report ignored links (e.g. mailto:...)")
	hideFailed    = flag.Bool("nofailed", false, "do NOT report failed links (e.g. 404)")
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
	checklinks.CrawlPage(pageURL, *timeout, *showSucceeded, *showIgnored, !*hideFailed)
}
