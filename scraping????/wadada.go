package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/proxy"
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
	var wg sync.WaitGroup
	for _, code := range lines {
		if code != "" {
			wg.Add(1)
			go func(code string) {
				defer wg.Done()
				if err := collyy(code); err != nil {
					log.Printf("Error scraping code %s: %v", code, err)
				}
			}(code)
		}
	}
	wg.Wait()
}

func collyy(code string) error {
	// Read proxy list from file
	proxies, err := readProxyList("proxylist.txt")
	if err != nil {
		return err
	}

	for _, proxyAddr := range proxies {
		// Create a SOCKS5 proxy dialer
		dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
		if err != nil {
			log.Printf("Failed to create proxy dialer for %s: %v", proxyAddr, err)
			continue // Skip to the next proxy on error
		}

		// Create a collector with proxy support
		c := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"),
			colly.Async(true), // Enable asynchronous requests
		)

		// Set dialer to the collector's Transport field
		c.WithTransport(&http.Transport{
			Dial: dialer.Dial,
		})

		url := "https://paste.fo/" + code

		// OnHTML callback function to parse the HTML and extract data
		c.OnHTML("textarea", func(e *colly.HTMLElement) {
			text := e.Text
			if text != "" {
				fmt.Printf("Non-empty Textarea:\n%s\n----------\n", text)
			}
		})

		// Error handling
		c.OnError(func(r *colly.Response, err error) {
			log.Printf("Request URL: %v failed with response: %v\nError: %v\n", r.Request.URL, r, err)
		})

		// Attempt to visit the URL with the current proxy
		err = c.Visit(url)
		if err == nil {
			// Successful request, break out of the loop
			break
		} else {
			log.Printf("Request with proxy %s failed: %v", proxyAddr, err)
		}
	}

	return nil
}

// readProxyList reads the proxy list from a file
func readProxyList(filename string) ([]string, error) {
	var proxies []string

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxies = append(proxies, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return proxies, nil
}
