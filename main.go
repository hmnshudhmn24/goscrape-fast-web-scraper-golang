// main.go
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gocolly/colly"
)

// ScrapedItem represents the structure of the data we want to collect
type ScrapedItem struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

var (
	outputFile = "output.json"
	csvFile    = "output.csv"
	targetURL  = "https://news.ycombinator.com/"
	scrapedData []ScrapedItem
	mutex      sync.Mutex
)

// saveToJSON saves scraped data to JSON file
func saveToJSON(data []ScrapedItem) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// saveToCSV saves scraped data to CSV file
func saveToCSV(data []ScrapedItem) error {
	file, err := os.Create(csvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Title", "Link"})
	for _, item := range data {
		writer.Write([]string{item.Title, item.Link})
	}
	return nil
}

// scrape performs the actual scraping using Colly
func scrape(wg *sync.WaitGroup, id int) {
	defer wg.Done()
	fmt.Printf("[Worker %d] Starting scrape task...\n", id)

	c := colly.NewCollector(
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5,
		Delay:       1 * time.Second,
	})

	c.OnHTML("a.storylink", func(e *colly.HTMLElement) {
		item := ScrapedItem{
			Title: e.Text,
			Link:  e.Attr("href"),
		}
		mutex.Lock()
		scrapedData = append(scrapedData, item)
		mutex.Unlock()
		fmt.Printf("[Worker %d] Scraped: %s\n", id, item.Title)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("[Worker %d] Visiting: %s\n", id, r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("[Worker %d] Error on %s: %v\n", id, r.Request.URL, err)
	})

	c.Visit(targetURL)
	c.Wait()

	fmt.Printf("[Worker %d] Finished scraping.\n", id)
}

func startWebServer() {
	http.HandleFunc("/results", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mutex.Lock()
		defer mutex.Unlock()
		json.NewEncoder(w).Encode(scrapedData)
	})

	fmt.Println("🌐 Web server running at http://localhost:8080/results")
	http.ListenAndServe(":8080", nil)
}

func main() {
	fmt.Println("🚀 GoScrape: Starting concurrent web scraper")

	start := time.Now()

	// Handle graceful shutdown
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\n💾 Saving scraped data before shutdown...")
		mutex.Lock()
		saveToJSON(scrapedData)
		saveToCSV(scrapedData)
		mutex.Unlock()
		done <- true
	}()

	// Start web server
	go startWebServer()

	// Start workers
	var wg sync.WaitGroup
	numWorkers := 3
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go scrape(&wg, i)
	}
	wg.Wait()

	// Save data after scraping
	mutex.Lock()
	saveToJSON(scrapedData)
	saveToCSV(scrapedData)
	mutex.Unlock()

	duration := time.Since(start)
	fmt.Printf("✅ Scraping completed in %s\n", duration)

	<-done
	fmt.Println("👋 Gracefully exited.")
}
    }
  ]
}
