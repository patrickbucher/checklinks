package checklinks

import (
	"bytes"
	"net/url"
	"testing"

	"golang.org/x/net/html"
)

const htmlDocument = `
<!DOCTYPE html>
<html>
	<head>
		<title>HTML Document</title>
	</head>
	<body>
		<div>
			<p><a href="https://github.com">github.com</a></p>
			<p><a href="https://paedubucher.ch">paedubucher.ch</a></p>
		</div>
	</body>
</html>
`

var hrefs = []string{"https://github.com", "https://paedubucher.ch"}

func TestExtractTagAttribute(t *testing.T) {
	data := bytes.NewBufferString(htmlDocument)
	root, _ := html.Parse(data)
	attributes := ExtractTagAttribute(root, "a", "href")
	if len(attributes) != len(hrefs) {
		t.Errorf("expected %d attributes, got %d", len(hrefs), len(attributes))
	} else if !isEqual(attributes, hrefs) {
		t.Errorf("expected a href attribute %v, got %v", hrefs, attributes)
	}
}

func isEqual[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

var qualifyURLTests = []struct {
	pageURL     string
	linkURL     string
	expectedURL string
}{
	{
		"https://paedubucher.ch/",
		"/articles/eat-more-cheese.html",
		"https://paedubucher.ch/articles/eat-more-cheese.html",
	},
	{
		"https://paedubucher.ch",
		"/articles/eat-more-cheese.html",
		"https://paedubucher.ch/articles/eat-more-cheese.html",
	},
	{
		"https://paedubucher.ch/articles/drink-more-milk/",
		"milk-manifesto.html",
		"https://paedubucher.ch/articles/drink-more-milk/milk-manifesto.html",
	},
	{
		"https://paedubucher.ch/articles/drink-more-milk",
		"milk-manifesto.html",
		"https://paedubucher.ch/articles/drink-more-milk/milk-manifesto.html",
	},
}

func TestQualifyInternalRootURL(t *testing.T) {
	for _, testCase := range qualifyURLTests {
		pageURL, _ := url.Parse(testCase.pageURL)
		linkURL, _ := url.Parse(testCase.linkURL)
		combinedURL := QualifyInternalURL(pageURL, linkURL)
		expectedURL := testCase.expectedURL
		if combinedURL.String() != expectedURL {
			t.Errorf("expected '%s', got '%s'", expectedURL, combinedURL.String())
		}
	}
}
