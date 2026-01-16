package gnewsdecoder

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

// GoogleDecoder is a struct that provides Google News URL decoding with optional proxy support.
type GoogleDecoder struct {
	client *http.Client
	proxy  string
}

// DecoderOption is a functional option for configuring GoogleDecoder
type DecoderOption func(*GoogleDecoder)

// WithProxy sets the proxy for the decoder.
// Supported formats:
//   - HTTP/HTTPS: http://user:pass@host:port or https://user:pass@host:port
//   - SOCKS5: socks5://user:pass@host:port
//   - Without auth: http://host:port or socks5://host:port
func WithProxy(proxyURL string) DecoderOption {
	return func(d *GoogleDecoder) {
		d.proxy = proxyURL
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) DecoderOption {
	return func(d *GoogleDecoder) {
		d.client = client
	}
}

// NewGoogleDecoder creates a new GoogleDecoder with optional configuration
func NewGoogleDecoder(opts ...DecoderOption) (*GoogleDecoder, error) {
	d := &GoogleDecoder{}

	for _, opt := range opts {
		opt(d)
	}

	if d.client == nil {
		d.client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	// Configure proxy if specified
	if d.proxy != "" {
		transport, err := createTransportWithProxy(d.proxy)
		if err != nil {
			return nil, err
		}
		d.client.Transport = transport
	}

	return d, nil
}

// createTransportWithProxy creates an HTTP transport with the specified proxy
func createTransportWithProxy(proxyURL string) (*http.Transport, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{}

	switch parsedURL.Scheme {
	case "socks5":
		// SOCKS5 proxy
		var auth *proxy.Auth
		if parsedURL.User != nil {
			auth = &proxy.Auth{
				User: parsedURL.User.Username(),
			}
			auth.Password, _ = parsedURL.User.Password()
		}
		dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}
		// Use Dial instead of DialContext for compatibility
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	default:
		// HTTP/HTTPS proxy
		transport.Proxy = http.ProxyURL(parsedURL)
	}

	return transport, nil
}

// GetBase64Str extracts the base64 string from a Google News URL
func (d *GoogleDecoder) GetBase64Str(sourceURL string) DecodeResult {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return DecodeResult{Status: false, Message: "failed to parse URL"}
	}

	path := splitPath(parsedURL.Path)
	if parsedURL.Host != "news.google.com" || len(path) <= 1 {
		return DecodeResult{Status: false, Message: "invalid Google News URL format"}
	}

	pathType := path[len(path)-2]
	if pathType != "articles" && pathType != "read" {
		return DecodeResult{Status: false, Message: "invalid Google News URL format"}
	}

	return DecodeResult{Status: true, DecodedURL: path[len(path)-1]}
}

// GetDecodingParams fetches signature and timestamp required for decoding
func (d *GoogleDecoder) GetDecodingParams(base64Str string) DecodingParams {
	return getDecodingParams(base64Str, d.client)
}

// DecodeURL decodes the Google News URL using the signature and timestamp
func (d *GoogleDecoder) DecodeURL(signature, timestamp, base64Str string) DecodeResult {
	return decodeURLWithParams(signature, timestamp, base64Str, d.client)
}

// Decode decodes a Google News article URL into its original source URL
func (d *GoogleDecoder) Decode(sourceURL string, interval *time.Duration) DecodeResult {
	return newDecoderV1WithClient(sourceURL, interval, d.client)
}

// splitPath splits a URL path into segments, removing empty strings
func splitPath(path string) []string {
	var result []string
	for _, p := range splitString(path, "/") {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

// ConcurrentDecoder provides concurrent URL decoding capabilities
type ConcurrentDecoder struct {
	decoder     *GoogleDecoder
	concurrency int
}

// NewConcurrentDecoder creates a new ConcurrentDecoder
func NewConcurrentDecoder(decoder *GoogleDecoder, concurrency int) *ConcurrentDecoder {
	if concurrency <= 0 {
		concurrency = 10
	}
	return &ConcurrentDecoder{
		decoder:     decoder,
		concurrency: concurrency,
	}
}

// DecodeResult with index for maintaining order
type indexedResult struct {
	index  int
	result DecodeResult
}

// DecodeURLs decodes multiple URLs concurrently
func (cd *ConcurrentDecoder) DecodeURLs(sourceURLs []string, interval *time.Duration) []DecodeResult {
	results := make([]DecodeResult, len(sourceURLs))

	// Create a channel to limit concurrency
	sem := make(chan struct{}, cd.concurrency)
	resultChan := make(chan indexedResult, len(sourceURLs))
	var wg sync.WaitGroup

	for i, sourceURL := range sourceURLs {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			result := cd.decoder.Decode(url, interval)
			resultChan <- indexedResult{index: idx, result: result}
		}(i, sourceURL)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for ir := range resultChan {
		results[ir.index] = ir.result
	}

	return results
}

// DecodeURLsWithContext decodes multiple URLs concurrently with context support for cancellation
func (cd *ConcurrentDecoder) DecodeURLsWithContext(ctx context.Context, sourceURLs []string, interval *time.Duration) []DecodeResult {
	results := make([]DecodeResult, len(sourceURLs))

	// Initialize all results with a default error
	for i := range results {
		results[i] = DecodeResult{Status: false, Message: "not processed"}
	}

	sem := make(chan struct{}, cd.concurrency)
	resultChan := make(chan indexedResult, len(sourceURLs))
	var wg sync.WaitGroup

	for i, sourceURL := range sourceURLs {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				resultChan <- indexedResult{
					index:  idx,
					result: DecodeResult{Status: false, Message: "context cancelled"},
				}
				return
			case sem <- struct{}{}:
				defer func() { <-sem }()
			}

			result := cd.decoder.Decode(url, interval)
			resultChan <- indexedResult{index: idx, result: result}
		}(i, sourceURL)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for ir := range resultChan {
		results[ir.index] = ir.result
	}

	return results
}

// GNewsDecoder is the main convenience function for decoding Google News URLs.
// This function creates a new decoder for each call, suitable for simple use cases.
//
// Parameters:
//   - sourceURL: The Google News article URL
//   - interval: Optional delay time before returning to avoid rate limits
//   - proxyURL: Optional proxy URL (supports http, https, socks5)
//
// Example:
//
//	result := gnewsdecoder.GNewsDecoder("https://news.google.com/...", nil, nil)
//	if result.Status {
//	    fmt.Println("Decoded URL:", result.DecodedURL)
//	}
func GNewsDecoder(sourceURL string, interval *time.Duration, proxyURL *string) DecodeResult {
	var opts []DecoderOption
	if proxyURL != nil && *proxyURL != "" {
		opts = append(opts, WithProxy(*proxyURL))
	}

	decoder, err := NewGoogleDecoder(opts...)
	if err != nil {
		return DecodeResult{Status: false, Message: err.Error()}
	}

	return decoder.Decode(sourceURL, interval)
}

// GNewsDecoderBatch decodes multiple Google News URLs using the efficient batch method
func GNewsDecoderBatch(sourceURLs []string) []DecodeResult {
	return DecoderV4(sourceURLs)
}

// GNewsDecoderConcurrent decodes multiple URLs concurrently with optional proxy support
func GNewsDecoderConcurrent(sourceURLs []string, concurrency int, interval *time.Duration, proxyURL *string) []DecodeResult {
	var opts []DecoderOption
	if proxyURL != nil && *proxyURL != "" {
		opts = append(opts, WithProxy(*proxyURL))
	}

	decoder, err := NewGoogleDecoder(opts...)
	if err != nil {
		results := make([]DecodeResult, len(sourceURLs))
		for i := range results {
			results[i] = DecodeResult{Status: false, Message: err.Error()}
		}
		return results
	}

	cd := NewConcurrentDecoder(decoder, concurrency)
	return cd.DecodeURLs(sourceURLs, interval)
}
