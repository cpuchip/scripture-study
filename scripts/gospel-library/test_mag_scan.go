package main

import (
    "context"
    "fmt"
    "time"

    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
    "github.com/cpuchip/scripture-study/scripts/gospel-library/internal/cache"
)

func main() {
    fmt.Println("Starting...")
    
    rawClient := api.NewClient("eng")
    fileCache := cache.New(".gospel-cache", "eng")
    cachedClient := cache.NewCachedClient(rawClient, fileCache)
    ctx := context.Background()
    
    fmt.Println("Testing API...")
    _, _, err := cachedClient.GetDynamic(ctx, "/ensign/2020/01")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Println("API works!")
    
    // Test the loop
    magazines := []struct {
        name      string
        startYear int
        endYear   int
    }{
        {"ensign", 1971, 2020},
        {"liahona", 1977, time.Now().Year() + 1},
    }
    
    months := []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"}
    
    var validIssues []string
    for _, mag := range magazines {
        fmt.Printf("Checking %s (%d-%d)...\n", mag.name, mag.startYear, mag.endYear)
        checked := 0
        for year := mag.startYear; year <= mag.endYear; year++ {
            for _, m := range months {
                checked++
                if checked % 100 == 0 {
                    fmt.Printf("  Checked %d issues...\n", checked)
                }
                uri := fmt.Sprintf("/%s/%d/%s", mag.name, year, m)
                _, _, err := cachedClient.GetDynamic(ctx, uri)
                if err == nil {
                    validIssues = append(validIssues, uri)
                }
            }
        }
        fmt.Printf("  Total checked: %d, found: %d valid\n", checked, len(validIssues))
    }
    
    fmt.Printf("\nTotal valid issues: %d\n", len(validIssues))
}
