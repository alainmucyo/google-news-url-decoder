//go:build ignore
// +build ignore

// This is an example program demonstrating how to use the gnewsdecoder package.
//
// Usage:
//
//	go run examples/main.go
package main

import (
	"fmt"
	"time"

	gnews "github.com/alainmucyo/gnewsdecoder"
)

func main() {
	// Example 1: Using the simple GNewsDecoder function
	fmt.Println("=== Example 1: Simple decode ===")
	sourceURL := "https://news.google.com/read/CBMi2AFBVV95cUxPd1ZCc1loODVVNHpnbFFTVHFkTG94eWh1NWhTeE9yT1RyNTRXMVV2S1VIUFM3ZlVkVjl6UHh3RkJ0bXdaTVRlcHBjMWFWTkhvZWVuM3pBMEtEdlllRDBveGdIUm9GUnJ4ajd1YWR5cWs3VFA5V2dsZnY1RDZhVDdORHRSSE9EalF2TndWdlh4bkJOWU5UMTdIV2RCc285Q2p3MFA4WnpodUNqN1RNREMwa3d5T2ZHS0JlX0MySGZLc01kWDNtUEkzemtkbWhTZXdQTmdfU1JJaXY?hl=en-US&gl=US&ceid=US%3Aen"

	result := gnews.GNewsDecoder(sourceURL, nil, nil)
	if result.Status {
		fmt.Println("Decoded URL:", result.DecodedURL)
	} else {
		fmt.Println("Error:", result.Message)
	}

	// Example 2: Using GNewsDecoder with interval
	fmt.Println("\n=== Example 2: Decode with interval ===")
	interval := 1 * time.Second
	result = gnews.GNewsDecoder(sourceURL, &interval, nil)
	if result.Status {
		fmt.Println("Decoded URL:", result.DecodedURL)
	} else {
		fmt.Println("Error:", result.Message)
	}

	// Example 3: Using GNewsDecoder with proxy
	fmt.Println("\n=== Example 3: Decode with proxy ===")
	proxy := "http://user:pass@localhost:8080"
	result = gnews.GNewsDecoder(sourceURL, nil, &proxy)
	if result.Status {
		fmt.Println("Decoded URL:", result.DecodedURL)
	} else {
		fmt.Println("Error:", result.Message)
	}

	// Example 4: Using GoogleDecoder class
	fmt.Println("\n=== Example 4: Using GoogleDecoder class ===")
	decoder, err := gnews.NewGoogleDecoder()
	if err != nil {
		fmt.Println("Failed to create decoder:", err)
		return
	}

	result = decoder.Decode(sourceURL, nil)
	if result.Status {
		fmt.Println("Decoded URL:", result.DecodedURL)
	} else {
		fmt.Println("Error:", result.Message)
	}

	// Example 5: Using GoogleDecoder with proxy
	fmt.Println("\n=== Example 5: GoogleDecoder with proxy ===")
	decoder, err = gnews.NewGoogleDecoder(gnews.WithProxy("socks5://localhost:1080"))
	if err != nil {
		fmt.Println("Failed to create decoder with proxy:", err)
		return
	}

	result = decoder.Decode(sourceURL, nil)
	if result.Status {
		fmt.Println("Decoded URL:", result.DecodedURL)
	} else {
		fmt.Println("Error:", result.Message)
	}

	// Example 6: Batch decoding multiple URLs
	fmt.Println("\n=== Example 6: Batch decode ===")
	sourceURLs := []string{
		"https://news.google.com/read/CBMilgFBVV95cUxOM0JJaFRwV2dqRDk5dEFpWmF1cC1IVml5WmVtbHZBRXBjZHBfaUsyalRpa1I3a2lKM1ZnZUI4MHhPU2sydi1nX3JrYU0xWjhLaHNfU0N6cEhOYVE2TEptRnRoZGVTU3kzZGJNQzc2aDZqYjJOR0xleTdsemdRVnJGLTVYTEhzWGw4Z19lR3AwR0F1bXlyZ0HSAYwBQVVfeXFMTXlLRDRJUFN5WHg3ZTI0X1F4SjN6bmFIck1IaGxFVVZyOFQxdk1JT3JUbl91SEhsU0NpQzkzRFdHSEtjVGhJNzY4ZTl6eXhESUQ3XzdWVTBGOGgwSmlXaVRmU3BsQlhPVjV4VWxET3FQVzJNbm5CUDlUOHJUTExaME5YbjZCX1NqOU9Ta3U?hl=en-US&gl=US&ceid=US%3Aen",
		"https://news.google.com/read/CBMiiAFBVV95cUxQOXZLdC1hSzFqQVVLWGJVZzlPaDYyNjdWTURScV9BbVp0SWhFNzZpSWZxSzdhc0tKbVlHMU13NmZVOFdidFFkajZPTm9SRnlZMWFRZ01CVHh0dXU0TjNVMUxZNk9Ibk5DV3hrYlRiZ20zYkIzSFhMQVVpcTFPc00xQjhhcGV1aXM00gF_QVVfeXFMTmtFQXMwMlY1el9WY0VRWEh5YkxXbHF0SjFLQVByNk1xS3hpdnBuUDVxOGZCQXl1QVFXaUVpbk5lUGgwRVVVT25tZlVUVWZqQzc4cm5MSVlfYmVlclFTOUFmTHF4eTlfemhTa2JKeG14bmNabENkSmZaeHB4WnZ5dw?hl=en-US&gl=US&ceid=US%3Aen",
	}

	results := gnews.GNewsDecoderBatch(sourceURLs)
	for i, r := range results {
		if r.Status {
			fmt.Printf("URL %d: %s\n", i+1, r.DecodedURL)
		} else {
			fmt.Printf("URL %d Error: %s\n", i+1, r.Message)
		}
	}

	// Example 7: Concurrent decoding
	fmt.Println("\n=== Example 7: Concurrent decode ===")
	results = gnews.GNewsDecoderConcurrent(sourceURLs, 5, nil, nil)
	for i, r := range results {
		if r.Status {
			fmt.Printf("URL %d: %s\n", i+1, r.DecodedURL)
		} else {
			fmt.Printf("URL %d Error: %s\n", i+1, r.Message)
		}
	}

	// Example 8: Using different decoder versions
	fmt.Println("\n=== Example 8: Different decoder versions ===")

	// DecoderV1 - Simple base64 decoding
	decoded := gnews.DecoderV1(sourceURL)
	fmt.Println("DecoderV1:", decoded)

	// DecoderV2 - With batch execute fallback
	decoded = gnews.DecoderV2(sourceURL)
	fmt.Println("DecoderV2:", decoded)

	// DecoderV3 - With proper error handling
	result = gnews.DecoderV3(sourceURL)
	if result.Status {
		fmt.Println("DecoderV3:", result.DecodedURL)
	} else {
		fmt.Println("DecoderV3 Error:", result.Message)
	}

	// NewDecoderV1 - Recommended decoder with signature/timestamp
	result = gnews.NewDecoderV1(sourceURL, nil)
	if result.Status {
		fmt.Println("NewDecoderV1:", result.DecodedURL)
	} else {
		fmt.Println("NewDecoderV1 Error:", result.Message)
	}

	fmt.Println("\n=== Done ===")
}
