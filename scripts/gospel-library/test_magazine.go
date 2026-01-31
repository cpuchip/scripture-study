package main

import (
    "context"
    "fmt"
    "encoding/json"

    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/cache"
)

func main() {
    rawClient := api.NewClient("eng")
    fileCache := cache.New(".gospel-cache", "eng")
    cachedClient := cache.NewCachedClient(rawClient, fileCache)
    ctx := context.Background()

    // Try the dynamic endpoint for a magazine issue
    fmt.Println("=== Testing /liahona/2026/01 (dynamic) ===")
    dynamic, _, err := cachedClient.GetDynamic(ctx, "/liahona/2026/01")
    if err != nil {
        fmt.Printf("Dynamic error: %v\n", err)
    } else if dynamic != nil {
        if dynamic.TOC != nil && len(dynamic.TOC.Entries) > 0 {
            fmt.Printf("TOC has %d entries\n", len(dynamic.TOC.Entries))
            for i, e := range dynamic.TOC.Entries {
                if i < 5 {
                    data, _ := json.MarshalIndent(e, "", "  ")
                    fmt.Printf("Entry %d:\n%s\n\n", i, string(data))
                }
            }
        }
        if dynamic.Collection != nil {
            fmt.Printf("Collection has %d sections\n", len(dynamic.Collection.Sections))
        }
    }
}
