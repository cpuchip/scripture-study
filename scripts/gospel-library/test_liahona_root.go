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

    // Try /liahona directly
    fmt.Println("=== Testing /liahona (collection) ===")
    coll, _, err := cachedClient.GetCollection(ctx, "/liahona")
    if err != nil {
        fmt.Printf("Collection error: %v\n", err)
    } else if coll != nil && len(coll.Sections) > 0 {
        fmt.Printf("Collection has %d sections\n", len(coll.Sections))
        for i, s := range coll.Sections {
            if i < 3 {
                fmt.Printf("Section %d: %s (%d entries)\n", i, s.Title, len(s.Entries))
                for j, e := range s.Entries {
                    if j < 5 {
                        fmt.Printf("  Entry: %s -> %s (type: %s)\n", e.Title, e.URI, e.Type)
                    }
                }
            }
        }
    }

    // Also try dynamic
    fmt.Println("\n=== Testing /liahona (dynamic) ===")
    dynamic, _, err := cachedClient.GetDynamic(ctx, "/liahona")
    if err != nil {
        fmt.Printf("Dynamic error: %v\n", err)
    } else if dynamic != nil {
        if dynamic.TOC != nil {
            fmt.Printf("TOC has %d entries\n", len(dynamic.TOC.Entries))
        }
        if dynamic.Collection != nil {
            fmt.Printf("Collection has %d sections\n", len(dynamic.Collection.Sections))
            for i, s := range dynamic.Collection.Sections {
                if i < 2 {
                    fmt.Printf("Section: %s (%d entries)\n", s.Title, len(s.Entries))
                    for j, e := range s.Entries {
                        if j < 5 {
                            data, _ := json.Marshal(e)
                            fmt.Printf("  %s\n", string(data))
                        }
                    }
                }
            }
        }
    }
}
