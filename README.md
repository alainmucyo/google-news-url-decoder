# Google News Decoder for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/alainmucyo/google-news-url-decoder.svg)](https://pkg.go.dev/github.com/alainmucyo/google-news-url-decoder)
[![Go Report Card](https://goreportcard.com/badge/github.com/alainmucyo/google-news-url-decoder)](https://goreportcard.com/report/github.com/alainmucyo/google-news-url-decoder)

A Go package to decode Google News article URLs to their original source URLs. This is a Go port of the Python package [googlenewsdecoder](https://github.com/SSujitX/google-news-url-decoder) by SSujitX.

## Features

- ✅ Multiple decoder versions (V1, V2, V3, V4, NewV1)
- ✅ Proxy support (HTTP, HTTPS, SOCKS5)
- ✅ Concurrent URL decoding with goroutines
- ✅ Batch decoding for multiple URLs
- ✅ Rate limiting/interval support
- ✅ Comprehensive error handling
- ✅ Context support for cancellation

## Installation

```bash
go get github.com/alainmucyo/google-news-url-decoder
```

## Quick Start

```go
package main

import (
    "fmt"
    gnews "github.com/alainmucyo/google-news-url-decoder"
)

func main() {
    sourceURL := "https://news.google.com/read/CBMi2AFBVV95cUxPd1ZCc1loODVVNHpnbFFTVHFkTG94eWh1NWhTeE9yT1RyNTRXMVV2S1VIUFM3ZlVkVjl6UHh3RkJ0bXdaTVRlcHBjMWFWTkhvZWVuM3pBMEtEdlllRDBveGdIUm9GUnJ4ajd1YWR5cWs3VFA5V2dsZnY1RDZhVDdORHRSSE9EalF2TndWdlh4bkJOWU5UMTdIV2RCc285Q2p3MFA4WnpodUNqN1RNREMwa3d5T2ZHS0JlX0MySGZLc01kWDNtUEkzemtkbWhTZXdQTmdfU1JJaXY?hl=en-US&gl=US&ceid=US%3Aen"
    
    result := gnews.GNewsDecoder(sourceURL, nil, nil)
    if result.Status {
        fmt.Println("Decoded URL:", result.DecodedURL)
    } else {
        fmt.Println("Error:", result.Message)
    }
}
```

## Usage

### Simple Decoding

```go
import gnews "github.com/alainmucyo/google-news-url-decoder"

// Basic usage
result := gnews.GNewsDecoder(sourceURL, nil, nil)
if result.Status {
    fmt.Println("Decoded URL:", result.DecodedURL)
}
```

### With Interval (Rate Limiting)

```go
import "time"

interval := 1 * time.Second
result := gnews.GNewsDecoder(sourceURL, &interval, nil)
```

### With Proxy

```go
// HTTP/HTTPS proxy
proxy := "http://user:pass@localhost:8080"
result := gnews.GNewsDecoder(sourceURL, nil, &proxy)

// SOCKS5 proxy
proxy := "socks5://user:pass@localhost:1080"
result := gnews.GNewsDecoder(sourceURL, nil, &proxy)
```

### Using GoogleDecoder Class

```go
// Create decoder
decoder, err := gnews.NewGoogleDecoder()
if err != nil {
    log.Fatal(err)
}

// Decode URL
result := decoder.Decode(sourceURL, nil)
```

### With Proxy Support

```go
decoder, err := gnews.NewGoogleDecoder(
    gnews.WithProxy("http://localhost:8080"),
)
if err != nil {
    log.Fatal(err)
}

result := decoder.Decode(sourceURL, nil)
```

### Batch Decoding

```go
urls := []string{
    "https://news.google.com/read/...",
    "https://news.google.com/read/...",
}

// Using batch decoder (single API request for multiple URLs)
results := gnews.GNewsDecoderBatch(urls)
for i, r := range results {
    if r.Status {
        fmt.Printf("URL %d: %s\n", i+1, r.DecodedURL)
    }
}
```

### Concurrent Decoding

```go
// Decode multiple URLs concurrently with 5 goroutines
results := gnews.GNewsDecoderConcurrent(urls, 5, nil, nil)
```

### Using ConcurrentDecoder with Context

```go
import "context"

decoder, _ := gnews.NewGoogleDecoder()
cd := gnews.NewConcurrentDecoder(decoder, 10)

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

results := cd.DecodeURLsWithContext(ctx, urls, nil)
```

## Decoder Versions

| Decoder | Description | Use Case |
|---------|-------------|----------|
| `DecoderV1` | Simple base64 decoding | Older Google News URLs |
| `DecoderV2` | With batch execute fallback | URLs with AU_yqL prefix |
| `DecoderV3` | With proper error handling | When you need status info |
| `DecoderV4` | Batch decode multiple URLs | Efficient batch processing |
| `NewDecoderV1` | Uses signature/timestamp method | **Recommended for most cases** |
| `GNewsDecoder` | Convenience function | Quick single URL decode |

## Supported Proxy Formats

- **HTTP/HTTPS**: `http://user:pass@host:port` or `http://host:port`
- **SOCKS5**: `socks5://user:pass@host:port` or `socks5://host:port`

## API Reference

### Types

```go
type DecodeResult struct {
    Status     bool   `json:"status"`
    DecodedURL string `json:"decoded_url,omitempty"`
    Message    string `json:"message,omitempty"`
}
```

### Functions

```go
// Simple decoders
func DecoderV1(sourceURL string) string
func DecoderV2(sourceURL string) string
func DecoderV3(sourceURL string) DecodeResult
func DecoderV4(sourceURLs []string) []DecodeResult
func NewDecoderV1(sourceURL string, interval *time.Duration) DecodeResult

// Convenience functions
func GNewsDecoder(sourceURL string, interval *time.Duration, proxyURL *string) DecodeResult
func GNewsDecoderBatch(sourceURLs []string) []DecodeResult
func GNewsDecoderConcurrent(sourceURLs []string, concurrency int, interval *time.Duration, proxyURL *string) []DecodeResult

// GoogleDecoder class
func NewGoogleDecoder(opts ...DecoderOption) (*GoogleDecoder, error)
func WithProxy(proxyURL string) DecoderOption
func WithHTTPClient(client *http.Client) DecoderOption

// ConcurrentDecoder
func NewConcurrentDecoder(decoder *GoogleDecoder, concurrency int) *ConcurrentDecoder
func (cd *ConcurrentDecoder) DecodeURLs(sourceURLs []string, interval *time.Duration) []DecodeResult
func (cd *ConcurrentDecoder) DecodeURLsWithContext(ctx context.Context, sourceURLs []string, interval *time.Duration) []DecodeResult
```

## Credits

- Original Python implementation by [SSujitX](https://github.com/SSujitX/google-news-url-decoder)
- Original script by [huksley](https://gist.github.com/huksley/)

## License

MIT License - see [LICENSE](LICENSE) for details.
