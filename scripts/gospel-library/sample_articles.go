package main

import (
    "context"
    "fmt"

    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/cache"
    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/tui"
)

func main() {
    rawClient := api.NewClient("eng")
    fileCache := cache.New(".gospel-cache", "eng")
    cachedClient := cache.NewCachedClient(rawClient, fileCache)
    ctx := context.Background()
    
    downloader := tui.NewDownloader(cachedClient, rawClient, "eng", "test-output")

    // Sample a few issues to see article counts
    issues := []string{
        "/liahona/2026/01",
        "/liahona/2020/01",
        "/liahona/2010/01",
        "/liahona/2000/01",
        "/ensign/2020/01",
        "/ensign/2010/01",
        "/ensign/2000/01",
    }
    
    total := 0
    for _, uri := range issues {
        uris, err := downloader.CrawlForContent(ctx, uri)
        if err != nil {
            fmt.Printf("%s: error - %v\n", uri, err)
        } else {
            fmt.Printf("%s: %d articles\n", uri, len(uris))
            total += len(uris)
        }
    }
    
    avg := float64(total) / float64(len(issues))
    fmt.Printf("\nAverage: %.1f articles per issue\n", avg)
    fmt.Printf("Estimated total for 1,136 issues: %.0f articles\n", avg * 1136)
}
