// Command gnewsdecoder is a CLI tool to decode Google News URLs to their original source URLs.
//
// Usage:
//
//	gnewsdecoder [flags] <url> [urls...]
//
// Example:
//
//	gnewsdecoder "https://news.google.com/read/CBMi..."
//	gnewsdecoder -proxy "http://localhost:8080" "https://news.google.com/read/CBMi..."
//	gnewsdecoder -batch "https://news.google.com/read/CBMi..." "https://news.google.com/read/CBMi..."
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	gnews "github.com/alainmucyo/gnewsdecoder"
)

func main() {
	// Flags
	proxyURL := flag.String("proxy", "", "Proxy URL (http://host:port or socks5://host:port)")
	intervalSec := flag.Int("interval", 0, "Interval in seconds between requests to avoid rate limits")
	batchMode := flag.Bool("batch", false, "Use batch mode for multiple URLs (more efficient)")
	concurrent := flag.Int("concurrent", 0, "Number of concurrent workers (0 = sequential)")
	jsonOutput := flag.Bool("json", false, "Output results as JSON")
	version := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Google News URL Decoder - Decode Google News URLs to original source URLs\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <url> [urls...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s \"https://news.google.com/read/CBMi...\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -proxy \"http://localhost:8080\" \"https://news.google.com/read/CBMi...\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -batch \"https://news.google.com/read/CBMi...\" \"https://news.google.com/read/CBMi...\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -concurrent 5 \"https://news.google.com/read/CBMi...\" \"https://news.google.com/read/CBMi...\"\n", os.Args[0])
	}

	flag.Parse()

	if *version {
		fmt.Printf("gnewsdecoder version %s\n", gnews.Version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Prepare interval
	var interval *time.Duration
	if *intervalSec > 0 {
		d := time.Duration(*intervalSec) * time.Second
		interval = &d
	}

	// Prepare proxy
	var proxy *string
	if *proxyURL != "" {
		proxy = proxyURL
	}

	var results []gnews.DecodeResult

	switch {
	case *batchMode && len(args) > 1:
		// Batch mode
		results = gnews.GNewsDecoderBatch(args)

	case *concurrent > 0:
		// Concurrent mode
		results = gnews.GNewsDecoderConcurrent(args, *concurrent, interval, proxy)

	default:
		// Sequential mode
		for _, url := range args {
			result := gnews.GNewsDecoder(url, interval, proxy)
			results = append(results, result)
		}
	}

	// Output results
	if *jsonOutput {
		outputJSON(results)
	} else {
		outputText(args, results)
	}
}

func outputJSON(results []gnews.DecodeResult) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func outputText(urls []string, results []gnews.DecodeResult) {
	exitCode := 0
	for i, result := range results {
		if result.Status {
			if len(results) > 1 {
				fmt.Printf("[%d] %s\n", i+1, result.DecodedURL)
			} else {
				fmt.Println(result.DecodedURL)
			}
		} else {
			if len(results) > 1 {
				fmt.Fprintf(os.Stderr, "[%d] Error: %s\n", i+1, result.Message)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s\n", result.Message)
			}
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}
