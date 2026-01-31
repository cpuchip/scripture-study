package main

import (
    "context"
    "fmt"
    "time"

    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/cache"
)

func main() {
    rawClient := api.NewClient("eng")
    fileCache := cache.New(".gospel-cache", "eng")
    cachedClient := cache.NewCachedClient(rawClient, fileCache)
    ctx := context.Background()

    magazines := []struct {
        name      string
        startYear int
    }{
        {"liahona", 1977},
        {"ensign", 1971},
    }
    
    months := []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"}
    currentYear := time.Now().Year() + 1  // +1 since they publish ahead
    
    var validIssues []string
    
    for _, mag := range magazines {
        fmt.Printf("=== Checking %s (%d-%d) ===\n", mag.name, mag.startYear, currentYear)
        count := 0
        for year := mag.startYear; year <= currentYear; year++ {
            for _, m := range months {
                uri := fmt.Sprintf("/%s/%d/%s", mag.name, year, m)
                _, _, err := cachedClient.GetDynamic(ctx, uri)
                if err == nil {
                    validIssues = append(validIssues, uri)
                    count++
                }
            }
        }
        fmt.Printf("Found %d valid issues for %s\n\n", count, mag.name)
    }
    
    fmt.Printf("\nTotal valid issues found: %d\n", len(validIssues))
    
    // Show first and last few
    fmt.Println("\nFirst 5:")
    for i := 0; i < 5 && i < len(validIssues); i++ {
        fmt.Println("  ", validIssues[i])
    }
    fmt.Println("\nLast 5:")
    for i := len(validIssues) - 5; i < len(validIssues); i++ {
        if i >= 0 {
            fmt.Println("  ", validIssues[i])
        }
    }
}
