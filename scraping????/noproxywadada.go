package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	charset := "abcdef1234567890"
	var number int
	fmt.Print("How many codes do you want? ")
	fmt.Scan(&number)

	var finalResult strings.Builder
	for j := 0; j < number; j++ {
		var result strings.Builder
		for i := 0; i < 12; i++ {
			index := rand.Intn(len(charset))
			result.WriteByte(charset[index])
		}
		finalResult.WriteString(result.String() + "\n")
	}
	fmt.Print(finalResult.String()) // Print all codes at once

	lines := strings.Split(finalResult.String(), "\n")
	fmt.Println("Reading each line:")

	// Use a WaitGroup to wait for all Goroutines to finish
	var wg sync.WaitGroup

	// Use a channel to receive errors from Goroutines
	errChan := make(chan error)

	// Use a semaphore to limit the number of concurrent Goroutines
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent Goroutines

	// Launch Goroutines to scrape concurrently
	for _, code := range lines {
		if code != "" {
			// Increment WaitGroup counter
			wg.Add(1)
			semaphore <- struct{}{} // Acquire semaphore
			go func(code string) {
				// Decrement WaitGroup counter when Goroutine finishes
				defer func() {
					wg.Done()
					<-semaphore // Release semaphore
				}()
				if err := collyy(code); err != nil {
					// Send error to error channel
					errChan <- err
				}
			}(code)
		}
	}

	// Close the error channel when all Goroutines finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Print errors received from Goroutines
	for err := range errChan {
		log.Println("Error scraping code:", err)
	}
}

func collyy(code string) error {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"),
	)
	url := "https://paste.fo/" + code

	// OnHTML callback function to parse the HTML and extract data
	c.OnHTML("textarea", func(e *colly.HTMLElement) {
		text := e.Text
		if text != "" {
			file, err := os.Create(code + ".txt")
			if err != nil {
				fmt.Printf("could not create file: %w", err)
			}
			defer file.Close()
			fmt.Printf("Non-empty Textarea:\n%s\n----------\n", text)
			if _, err := file.WriteString(text + "\n----------\n"); err != nil {
				fmt.Println("Could not write to file:", err)
			}
		}
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %v failed with response: %v\nError: %v\n", r.Request.URL, r, err)
	})

	// Visit the URL to fetch the HTML content
	return c.Visit(url)
}
