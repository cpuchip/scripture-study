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

    // Try crawling just /liahona/2026 (one year)
    fmt.Println("=== Crawling /liahona/2026 (one year) ===")
    uris, err := downloader.CrawlForContent(ctx, "/liahona/2026")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Found %d content URIs for Liahona 2026\n\n", len(uris))

    // Try crawling just /ensign/2020 (one year)
    fmt.Println("=== Crawling /ensign/2020 (one year) ===")
    uris2, err := downloader.CrawlForContent(ctx, "/ensign/2020")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Found %d content URIs for Ensign 2020\n", len(uris2))
}
