package exercise

import (
	"fmt"
	"sync"
)

// Fetcher defines the interface to load an URL and return its body together with embeded urls
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type crawledUrls struct {
	mu   sync.Mutex
	urls map[string]bool
}

func hasProcessed(url string, crawled *crawledUrls) bool {
	crawled.mu.Lock()
	defer crawled.mu.Unlock()

	processed := false
	if _, ok := crawled.urls[url]; ok {
		processed = true
	} else {
		crawled.urls[url] = true
	}

	return processed
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, crawled *crawledUrls, wg *sync.WaitGroup) {
	defer wg.Done()

	if depth <= 0 {
		return
	}

	if processed := hasProcessed(url, crawled); processed {
		fmt.Printf("%s already fetched\n", url)
		return
	}

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		wg.Add(1)
		go Crawl(u, depth-1, fetcher, crawled, wg)
	}
}

// CrawlerMain implements the web crawler
func CrawlerMain() {
	crawled := crawledUrls{
		urls: make(map[string]bool),
	}
	var wg sync.WaitGroup

	wg.Add(1)
	go Crawl("https://golang.org/", 4, fetcher, &crawled, &wg)

	// wait for all goroutines to finish
	wg.Wait()
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
