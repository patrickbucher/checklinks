# `checklinks`: Crawl a Website for Dead URLs

The `checklinks` utility takes a single website address and crawls that page for
links (i.e. `href` attributes of `<a>` tags).

Internal links (pointing to another page on the same site) are checked using a
`GET` request. (TODO: The whole process shall be repeated recursively for the
page received.)

External links (pointing to another domain) are checked using a `HEAD` request
to save traffic. (Those links are not followed any further.)

The success and failure of each individual link is reported to the terminal.
(TODO: Define flags to control the output; by default, only broken links and the
page they're on shall be reported; working links should only be reported if a
`-verbose` flag is set.)

## Run It

    $ go run cmd/checklinks.go [url]

If the URL does not start with an `http://` or `https://` prefix, `http://` is
automatically assumed.

## Build It, Then Run It

    $ go build -o checklinks cmd/main.go
    $ ./checklinks [url]
