# `checklinks`: Crawl a Website for Dead URLs

The `checklinks` utility takes a single website address and crawls that page for
links (i.e. `href` attributes of `<a>` tags). TLS issues are ignored.

## Run It

    $ go run cmd/checklinks.go [url]

If the URL does not start with an `http://` or `https://` prefix, `http://` is
automatically assumed.

## Build It, Then Run It

    $ go build cmd/checklinks.go
    $ ./checklinks [url]

## Install It

Pick a tag (e.g. `v0.0.8`) and use `go install` to install that particular
version:

    $ go install github.com/patrickbucher/checklinks/checklinks@v0.0.8
    go: downloading github.com/patrickbucher/checklinks v0.0.8
    go: downloading golang.org/x/net v0.0.0-20220412020605-290c469a71a5

## Flags

The success and failure of each individual link is reported to the terminal. Use
the flags to control the output and request timeout:

    $ ./checklinks -help
    Usage of ./checklinks:
      -ignored
            report ignored links (e.g. mailto:...)
      -nofailed
            do NOT report failed links (e.g. 404)
      -success
            report succeeded links (OK)
      -timeout int
            request timeout (in seconds) (default 10)

## TODO

- [ ] introduce command line flags
    - [ ] user agent (optional)
    - [ ] level of parallelism (optional)
    - [ ] allow insecure SSL/TLS
- [ ] refactor code
    - [ ] introduce Config struct for handing over the entire configuration
      from the command line to the crawler function
    - [ ] introduce Channels struct for handing over channels to Process functions
