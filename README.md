# `checklinks`: Crawl a Website for Dead URLs

The `checklinks` utility takes a single website address and crawls that page for
links (i.e. `href` attributes of `<a>` tags).

Internal links (pointing to another page on the same site) are checked using a
`GET` request.

External links (pointing to another domain) are checked using a `HEAD` request
to save traffic. (Those links are not followed any further.)

## Run It

    $ go run cmd/checklinks.go [url]

If the URL does not start with an `http://` or `https://` prefix, `http://` is
automatically assumed.

## Build It, Then Run It

    $ go build cmd/checklinks.go
    $ ./checklinks [url]

## Install It

Pick a tag (e.g. `v0.0.2`) and use `go install` to install that particular
version:

    $ go install github.com/patrickbucher/checklinks/cmd@v0.0.2
    go: downloading github.com/patrickbucher/checklinks v0.0.2

## Flags

The success and failure of each individual link is reported to the terminal. Use
the flags to control the output and request timeout:

    $ ./checklinks -help
    Usage of ./checklinks:
      -failed
            report failed links (e.g. 404) (default true)
      -ignored
            report ignored links (e.g. mailto:...)
      -success
            report succeeded links (OK)
      -timeout int
            request timeout (in seconds) (default 10)
