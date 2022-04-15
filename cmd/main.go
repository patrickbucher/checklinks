package main

import (
	"flag"
	"fmt"
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
	url := args[0]
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	doc, err := checklinks.FetchDocument(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch document at %s: %v", url, err)
		os.Exit(1)
	}
	fmt.Println(doc)
}
