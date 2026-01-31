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
    downloader := tui.NewDownloader(cachedClient, rawClient, "eng", "gospel-library")
    ctx := context.Background()

    fmt.Println("=== Crawling /liahona/2026/01 ===")
    uris, err := downloader.CrawlForContent(ctx, "/liahona/2026/01")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Found %d content URIs:\n", len(uris))
    for i, uri := range uris {
        if i < 30 {
            fmt.Printf("  %s\n", uri)
        } else {
            fmt.Printf("  ... and %d more\n", len(uris)-30)
            break
        }
    }
}
