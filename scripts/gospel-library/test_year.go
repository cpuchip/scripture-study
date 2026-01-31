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

    fmt.Println("=== Testing /liahona/2026 (dynamic) ===")
    dynamic, _, err := cachedClient.GetDynamic(ctx, "/liahona/2026")
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
        } else {
            fmt.Println("No TOC")
        }
        if dynamic.Collection != nil && len(dynamic.Collection.Sections) > 0 {
            fmt.Printf("Collection has %d sections\n", len(dynamic.Collection.Sections))
            for i, s := range dynamic.Collection.Sections {
                if i < 3 {
                    fmt.Printf("Section %d: %s (%d entries)\n", i, s.Title, len(s.Entries))
                    for j, e := range s.Entries {
                        if j < 3 {
                            fmt.Printf("  Entry: %s -> %s\n", e.Title, e.URI)
                        }
                    }
                }
            }
        } else {
            fmt.Println("No Collection")
        }
    }
}
