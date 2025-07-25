package main

import (
    "fmt"
    "log"
    "os"
    "sync"
    "time"

    "github.com/gocolly/colly"
    "github.com/joho/godotenv"
    "goscrape/database"
    "goscrape/scraper"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found. Using default settings.")
    }

    db := database.InitDB()
    defer db.Close()

    targets := []string{
        "https://news.ycombinator.com/",
        "https://techcrunch.com/",
    }

    var wg sync.WaitGroup

    for _, url := range targets {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            scraper.Scrape(u, db)
        }(url)
    }

    wg.Wait()
    fmt.Println("Scraping completed at", time.Now())
}