package main

import (
    "context"
    "fmt"

    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/cache"
)

func main() {
    rawClient := api.NewClient("eng")
    fileCache := cache.New(".gospel-cache", "eng")
    cachedClient := cache.NewCachedClient(rawClient, fileCache)
    ctx := context.Background()

    // Check which 2026 months exist
    months := []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"}
    
    fmt.Println("=== Checking Liahona 2026 months ===")
    for _, m := range months {
        uri := fmt.Sprintf("/liahona/2026/%s", m)
        _, _, err := cachedClient.GetDynamic(ctx, uri)
        if err == nil {
            fmt.Printf("✓ %s exists\n", uri)
        }
    }
    
    fmt.Println("\n=== Checking Liahona 2025 months ===")
    for _, m := range months {
        uri := fmt.Sprintf("/liahona/2025/%s", m)
        _, _, err := cachedClient.GetDynamic(ctx, uri)
        if err == nil {
            fmt.Printf("✓ %s exists\n", uri)
        }
    }
}
