package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/tour/tree"
)

func walkRecurse(t *tree.Tree, ch chan int) {
	if t.Left != nil {
		walkRecurse(t.Left, ch)
	}
	ch <- t.Value
	if t.Right != nil {
		walkRecurse(t.Right, ch)
	}
}

func Walk(t *tree.Tree, ch chan int) {
	walkRecurse(t, ch)
	close(ch)
}

func Same(t1, t2 *tree.Tree) bool {
	ch1 := make(chan int)
	ch2 := make(chan int)
	go Walk(t1, ch1)
	go Walk(t2, ch2)
	/*
		for i := range ch1 {
			fmt.Println(i)
		}
		for j := range ch2 {
			fmt.Println(j)
		}
		return true
	*/
	for k := range ch1 {
		select {
		case g := <-ch2:
			if k != g {
				return false
			}
		default:
		}
	}
	return true
}

// SafeCounter is safe to use concurrently.
type SafeCounter struct {
	mu sync.Mutex
	v  map[string]int
}

// Inc increments the counter for the given key.
func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.v[key]++
	c.mu.Unlock()
}

// Value returns the current value of the counter for the given key.
func (c *SafeCounter) Value(key string) int {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mu.Unlock()
	return c.v[key]
}

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, safer *SafeCounter) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	//safer.Inc(url)
	fmt.Println(safer.Value(url))

	//fmt.Printf("found: %s %q\n", url, body)
	if safer.Value(url) <= 0 {
		fmt.Printf("found: %s %q\n", url, body)
		safer.Inc(url)
		for _, u := range urls {
			go Crawl(u, depth-1, fetcher, safer)
		}
	}
	//for _, u := range urls {
	//	Crawl(u, depth-1, fetcher, safer)
	//}
}

func WebCrawlTest() {
	c := SafeCounter{v: make(map[string]int)}
	Crawl("https://golang.org/", 4, fetcher, &c)
	time.Sleep(10 * time.Second)
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

func main() {
	fmt.Println("Binary Tree Checker")
	fmt.Println(Same(tree.New(1), tree.New(1)))
	fmt.Println(Same(tree.New(1), tree.New(3)))
	WebCrawlTest()
}
