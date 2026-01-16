package gnewsdecoder_test

import (
	"testing"
	"time"

	gnews "github.com/alainmucyo/google-news-url-decoder"
)

func TestDecoderV1_SimpleURL(t *testing.T) {
	// This is a test URL format - in real usage, this would be an actual Google News URL
	sourceURL := "https://news.google.com/rss/articles/CBMiLmh0dHBzOi8vd3d3LmJiYy5jb20vbmV3cy9hcnRpY2xlcy9jampqbnhkdjE4OG_SATJodHRwczovL3d3dy5iYmMuY29tL25ld3MvYXJ0aWNsZXMvY2pqam54ZHYxODhvLmFtcA?oc=5"

	result := gnews.DecoderV1(sourceURL)

	// The result should either be the decoded URL or the original if decoding fails
	if result == "" {
		t.Error("DecoderV1 returned empty string")
	}
}

func TestDecoderV3_InvalidURL(t *testing.T) {
	result := gnews.DecoderV3("https://example.com/not-a-google-news-url")

	if result.Status {
		t.Error("Expected Status to be false for invalid URL")
	}

	if result.Message == "" {
		t.Error("Expected error message for invalid URL")
	}
}

func TestDecoderV3_MalformedURL(t *testing.T) {
	result := gnews.DecoderV3("not-a-valid-url-at-all")

	if result.Status {
		t.Error("Expected Status to be false for malformed URL")
	}
}

func TestDecoderV4_BatchDecode(t *testing.T) {
	sourceURLs := []string{
		"https://example.com/not-google-news",
		"https://news.google.com/invalid/path",
	}

	results := gnews.DecoderV4(sourceURLs)

	if len(results) != len(sourceURLs) {
		t.Errorf("Expected %d results, got %d", len(sourceURLs), len(results))
	}

	// Both should fail as they're invalid URLs
	for i, result := range results {
		if result.Status {
			t.Errorf("Expected result %d to have Status false", i)
		}
	}
}

func TestGoogleDecoder_Creation(t *testing.T) {
	decoder, err := gnews.NewGoogleDecoder()
	if err != nil {
		t.Fatalf("Failed to create GoogleDecoder: %v", err)
	}

	if decoder == nil {
		t.Fatal("GoogleDecoder is nil")
	}
}

func TestGoogleDecoder_WithProxy(t *testing.T) {
	decoder, err := gnews.NewGoogleDecoder(gnews.WithProxy("http://localhost:8080"))
	if err != nil {
		t.Fatalf("Failed to create GoogleDecoder with proxy: %v", err)
	}

	if decoder == nil {
		t.Fatal("GoogleDecoder is nil")
	}
}

func TestGoogleDecoder_InvalidURL(t *testing.T) {
	decoder, err := gnews.NewGoogleDecoder()
	if err != nil {
		t.Fatalf("Failed to create GoogleDecoder: %v", err)
	}

	result := decoder.Decode("https://example.com/not-google-news", nil)
	if result.Status {
		t.Error("Expected Status to be false for invalid URL")
	}
}

func TestGoogleDecoder_GetBase64Str(t *testing.T) {
	decoder, _ := gnews.NewGoogleDecoder()

	tests := []struct {
		name      string
		url       string
		wantOK    bool
		wantEmpty bool
	}{
		{
			name:      "Valid articles URL",
			url:       "https://news.google.com/articles/CBMitest123",
			wantOK:    true,
			wantEmpty: false,
		},
		{
			name:      "Valid read URL",
			url:       "https://news.google.com/read/CBMitest456",
			wantOK:    true,
			wantEmpty: false,
		},
		{
			name:      "Invalid URL",
			url:       "https://example.com/some/path",
			wantOK:    false,
			wantEmpty: true,
		},
		{
			name:      "Invalid path structure",
			url:       "https://news.google.com/invalid/path/structure",
			wantOK:    false,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decoder.GetBase64Str(tt.url)
			if result.Status != tt.wantOK {
				t.Errorf("GetBase64Str() status = %v, want %v", result.Status, tt.wantOK)
			}
			if tt.wantEmpty && result.DecodedURL != "" {
				t.Errorf("GetBase64Str() decoded_url should be empty for invalid URLs")
			}
		})
	}
}

func TestConcurrentDecoder_Creation(t *testing.T) {
	decoder, _ := gnews.NewGoogleDecoder()
	cd := gnews.NewConcurrentDecoder(decoder, 5)

	if cd == nil {
		t.Fatal("ConcurrentDecoder is nil")
	}
}

func TestConcurrentDecoder_DefaultConcurrency(t *testing.T) {
	decoder, _ := gnews.NewGoogleDecoder()
	cd := gnews.NewConcurrentDecoder(decoder, 0) // 0 should default to 10

	if cd == nil {
		t.Fatal("ConcurrentDecoder is nil with default concurrency")
	}
}

func TestConcurrentDecoder_DecodeInvalidURLs(t *testing.T) {
	decoder, _ := gnews.NewGoogleDecoder()
	cd := gnews.NewConcurrentDecoder(decoder, 3)

	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
	}

	results := cd.DecodeURLs(urls, nil)

	if len(results) != len(urls) {
		t.Errorf("Expected %d results, got %d", len(urls), len(results))
	}

	for i, result := range results {
		if result.Status {
			t.Errorf("Result %d: expected Status false for invalid URL", i)
		}
	}
}

func TestGNewsDecoder_Convenience(t *testing.T) {
	result := gnews.GNewsDecoder("https://example.com/invalid", nil, nil)

	if result.Status {
		t.Error("Expected Status to be false for invalid URL")
	}
}

func TestGNewsDecoderBatch_Convenience(t *testing.T) {
	urls := []string{"https://example.com/1", "https://example.com/2"}
	results := gnews.GNewsDecoderBatch(urls)

	if len(results) != len(urls) {
		t.Errorf("Expected %d results, got %d", len(urls), len(results))
	}
}

func TestGNewsDecoderConcurrent_Convenience(t *testing.T) {
	urls := []string{"https://example.com/1", "https://example.com/2"}
	results := gnews.GNewsDecoderConcurrent(urls, 2, nil, nil)

	if len(results) != len(urls) {
		t.Errorf("Expected %d results, got %d", len(urls), len(results))
	}
}

func TestGNewsDecoderConcurrent_WithProxy(t *testing.T) {
	urls := []string{"https://example.com/1"}
	proxy := "http://localhost:8080"
	results := gnews.GNewsDecoderConcurrent(urls, 1, nil, &proxy)

	if len(results) != len(urls) {
		t.Errorf("Expected %d results, got %d", len(urls), len(results))
	}
}

func TestNewDecoderV1_InvalidURL(t *testing.T) {
	result := gnews.NewDecoderV1("https://example.com/not-google-news", nil)

	if result.Status {
		t.Error("Expected Status to be false for invalid URL")
	}
}

func TestNewDecoderV1_WithInterval(t *testing.T) {
	interval := 100 * time.Millisecond

	start := time.Now()
	_ = gnews.NewDecoderV1("https://example.com/not-google-news", &interval)
	elapsed := time.Since(start)

	// The function should fail before applying the interval since URL is invalid
	// Just ensure it doesn't panic
	if elapsed < 0 {
		t.Error("Time should not be negative")
	}
}

// Benchmark tests
func BenchmarkDecoderV1(b *testing.B) {
	url := "https://news.google.com/rss/articles/CBMiLmh0dHBzOi8vd3d3LmJiYy5jb20vbmV3cy9hcnRpY2xlcy9jampqbnhkdjE4OG_SATJodHRwczovL3d3dy5iYmMuY29tL25ld3MvYXJ0aWNsZXMvY2pqam54ZHYxODhvLmFtcA?oc=5"

	for i := 0; i < b.N; i++ {
		gnews.DecoderV1(url)
	}
}

func BenchmarkDecoderV3(b *testing.B) {
	url := "https://example.com/invalid"

	for i := 0; i < b.N; i++ {
		gnews.DecoderV3(url)
	}
}
