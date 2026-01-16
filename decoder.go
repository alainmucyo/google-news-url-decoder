// Package gnewsdecoder provides functions to decode Google News URLs to their original source URLs.
// This is a Go port of the Python package googlenewsdecoder by SSujitX.
package gnewsdecoder

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Version of the package
const Version = "0.1.0"

// DecodeResult represents the result of a URL decoding operation
type DecodeResult struct {
	Status     bool   `json:"status"`
	DecodedURL string `json:"decoded_url,omitempty"`
	Message    string `json:"message,omitempty"`
}

// DecodingParams contains the parameters needed for decoding
type DecodingParams struct {
	Status    bool   `json:"status"`
	Signature string `json:"signature,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Base64Str string `json:"base64_str,omitempty"`
	Message   string `json:"message,omitempty"`
}

// BatchDecodeResult represents the result of batch URL decoding
type BatchDecodeResult struct {
	Status bool     `json:"status"`
	URLs   []string `json:"urls,omitempty"`
	Error  string   `json:"error,omitempty"`
}

// DecoderV1 decodes Google News URLs using base64 decoding (simple method).
// This works for older/simpler Google News URL formats.
func DecoderV1(sourceURL string) string {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return sourceURL
	}

	path := strings.Split(parsedURL.Path, "/")
	if parsedURL.Host == "news.google.com" && len(path) > 1 && path[len(path)-2] == "articles" {
		base64Str := path[len(path)-1]
		decodedBytes, err := base64.URLEncoding.DecodeString(base64Str + "==")
		if err != nil {
			return sourceURL
		}
		decodedStr := string(decodedBytes)

		// Remove prefix
		prefix := string([]byte{0x08, 0x13, 0x22})
		if strings.HasPrefix(decodedStr, prefix) {
			decodedStr = decodedStr[len(prefix):]
		}

		// Remove suffix
		suffix := string([]byte{0xD2, 0x01, 0x00})
		if strings.HasSuffix(decodedStr, suffix) {
			decodedStr = decodedStr[:len(decodedStr)-len(suffix)]
		}

		// Extract URL based on length byte
		if len(decodedStr) > 0 {
			length := int(decodedStr[0])
			if length >= 0x80 && len(decodedStr) > length+1 {
				decodedStr = decodedStr[2 : length+1]
			} else if len(decodedStr) > length+1 {
				decodedStr = decodedStr[1 : length+1]
			}
		}

		return decodedStr
	}

	return sourceURL
}

// fetchDecodedBatchExecute fetches the decoded URL using Google's batch execute API
func fetchDecodedBatchExecute(id string, client *http.Client) (string, error) {
	s := fmt.Sprintf(
		`[[["Fbv4je","[\"garturlreq\",[[\"en-US\",\"US\",[\"FINANCE_TOP_INDICES\",\"WEB_TEST_1_0_0\"],`+
			`null,null,1,1,\"US:en\",null,180,null,null,null,null,null,0,null,null,[1608992183,723341000]],`+
			`\"en-US\",\"US\",1,[2,3,4,8],1,0,\"655000234\",0,0,null,0],\"%s\"]",null,"generic"]]]`,
		id,
	)

	reqBody := url.Values{}
	reqBody.Set("f.req", s)

	req, err := http.NewRequest("POST", "https://news.google.com/_/DotsSplashUi/data/batchexecute?rpcids=Fbv4je", strings.NewReader(reqBody.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	req.Header.Set("Referer", "https://news.google.com/")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch data from Google, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	text := string(body)
	header := `[\"garturlres\",\"`
	footer := `\",`

	if !strings.Contains(text, header) {
		return "", fmt.Errorf("header not found in response")
	}

	parts := strings.SplitN(text, header, 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("failed to parse response")
	}

	start := parts[1]
	if !strings.Contains(start, footer) {
		return "", fmt.Errorf("footer not found in response")
	}

	urlParts := strings.SplitN(start, footer, 2)
	return urlParts[0], nil
}

// DecoderV2 decodes Google News URLs with batch execute fallback for AU_yqL prefixed URLs.
// Returns the decoded URL or the original URL if decoding fails.
func DecoderV2(sourceURL string) string {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return sourceURL
	}

	path := strings.Split(parsedURL.Path, "/")
	if parsedURL.Host == "news.google.com" && len(path) > 1 && (path[len(path)-2] == "articles" || path[len(path)-2] == "read") {
		base64Str := path[len(path)-1]
		decodedBytes, err := base64.URLEncoding.DecodeString(base64Str + "==")
		if err != nil {
			return sourceURL
		}
		decodedStr := string(decodedBytes)

		// Remove prefix
		prefix := string([]byte{0x08, 0x13, 0x22})
		if strings.HasPrefix(decodedStr, prefix) {
			decodedStr = decodedStr[len(prefix):]
		}

		// Remove suffix
		suffix := string([]byte{0xD2, 0x01, 0x00})
		if strings.HasSuffix(decodedStr, suffix) {
			decodedStr = decodedStr[:len(decodedStr)-len(suffix)]
		}

		// Extract URL based on length byte
		if len(decodedStr) > 0 {
			length := int(decodedStr[0])
			if length >= 0x80 && len(decodedStr) > length+1 {
				decodedStr = decodedStr[2 : length+1]
			} else if len(decodedStr) > length+1 {
				decodedStr = decodedStr[1 : length+1]
			}
		}

		// If URL starts with AU_yqL, use batch execute
		if strings.HasPrefix(decodedStr, "AU_yqL") {
			client := &http.Client{Timeout: 30 * time.Second}
			decoded, err := fetchDecodedBatchExecute(base64Str, client)
			if err != nil {
				return sourceURL
			}
			return decoded
		}

		return decodedStr
	}

	return sourceURL
}

// DecoderV3 decodes Google News URLs with proper error handling and status reporting.
// Returns a DecodeResult with status and decoded URL or error message.
func DecoderV3(sourceURL string) DecodeResult {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return DecodeResult{Status: false, Message: fmt.Sprintf("failed to parse URL: %v", err)}
	}

	path := strings.Split(parsedURL.Path, "/")
	if parsedURL.Host == "news.google.com" && len(path) > 1 && (path[len(path)-2] == "articles" || path[len(path)-2] == "read") {
		base64Str := path[len(path)-1]
		decodedBytes, err := base64.URLEncoding.DecodeString(base64Str + "==")
		if err != nil {
			return DecodeResult{Status: false, Message: fmt.Sprintf("failed to decode base64: %v", err)}
		}
		decodedStr := string(decodedBytes)

		// Remove prefix
		prefix := string([]byte{0x08, 0x13, 0x22})
		if strings.HasPrefix(decodedStr, prefix) {
			decodedStr = decodedStr[len(prefix):]
		}

		// Remove suffix
		suffix := string([]byte{0xD2, 0x01, 0x00})
		if strings.HasSuffix(decodedStr, suffix) {
			decodedStr = decodedStr[:len(decodedStr)-len(suffix)]
		}

		// Extract URL based on length byte
		if len(decodedStr) > 0 {
			length := int(decodedStr[0])
			if length >= 0x80 && len(decodedStr) > length+1 {
				decodedStr = decodedStr[2 : length+1]
			} else if len(decodedStr) > length+1 {
				decodedStr = decodedStr[1 : length+1]
			}
		}

		// If URL starts with AU_yqL, use batch execute
		if strings.HasPrefix(decodedStr, "AU_yqL") {
			client := &http.Client{Timeout: 30 * time.Second}
			decoded, err := fetchDecodedBatchExecute(base64Str, client)
			if err != nil {
				return DecodeResult{Status: false, Message: fmt.Sprintf("batch execute failed: %v", err)}
			}
			return DecodeResult{Status: true, DecodedURL: decoded}
		}

		return DecodeResult{Status: true, DecodedURL: decodedStr}
	}

	return DecodeResult{Status: false, Message: "invalid Google News URL"}
}

// fetchDecodedBatchExecuteMultiple fetches multiple decoded URLs in a single batch request
func fetchDecodedBatchExecuteMultiple(ids []string, client *http.Client) (BatchDecodeResult, error) {
	var envelopes []string
	for i, id := range ids {
		envelope := fmt.Sprintf(
			`["Fbv4je","[\"garturlreq\",[[\"en-US\",\"US\",[\"FINANCE_TOP_INDICES\",\"WEB_TEST_1_0_0\"],`+
				`null,null,1,1,\"US:en\",null,180,null,null,null,null,null,0,null,null,[1608992183,723341000]],`+
				`\"en-US\",\"US\",1,[2,3,4,8],1,0,\"655000234\",0,0,null,0],\"%s\"]",null,"%d"]`,
			id, i+1,
		)
		envelopes = append(envelopes, envelope)
	}

	s := fmt.Sprintf("[[%s]]", strings.Join(envelopes, ","))

	reqBody := url.Values{}
	reqBody.Set("f.req", s)

	req, err := http.NewRequest("POST", "https://news.google.com/_/DotsSplashUi/data/batchexecute?rpcids=Fbv4je", strings.NewReader(reqBody.Encode()))
	if err != nil {
		return BatchDecodeResult{Status: false, Error: err.Error()}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	req.Header.Set("Referer", "https://news.google.com/")

	resp, err := client.Do(req)
	if err != nil {
		return BatchDecodeResult{Status: false, Error: err.Error()}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errMsg := fmt.Sprintf("failed to fetch data from Google, status: %d", resp.StatusCode)
		return BatchDecodeResult{Status: false, Error: errMsg}, errors.New(errMsg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return BatchDecodeResult{Status: false, Error: err.Error()}, err
	}

	text := string(body)
	header := `[\"garturlres\",\"`
	footer := `\",`

	var urls []string
	for strings.Contains(text, header) {
		parts := strings.SplitN(text, header, 2)
		if len(parts) < 2 {
			break
		}
		start := parts[1]
		if !strings.Contains(start, footer) {
			break
		}
		urlParts := strings.SplitN(start, footer, 2)
		urls = append(urls, urlParts[0])
		text = urlParts[1]
	}

	return BatchDecodeResult{Status: true, URLs: urls}, nil
}

// DecoderV4 decodes multiple Google News URLs in batch.
// This is more efficient when decoding multiple URLs as it batches API requests.
func DecoderV4(sourceURLs []string) []DecodeResult {
	results := make([]DecodeResult, len(sourceURLs))
	batchIDs := make([]string, 0)
	idToIndex := make(map[string]int)

	client := &http.Client{Timeout: 30 * time.Second}

	for i, sourceURL := range sourceURLs {
		parsedURL, err := url.Parse(sourceURL)
		if err != nil {
			results[i] = DecodeResult{Status: false, Message: fmt.Sprintf("failed to parse URL: %v", err)}
			continue
		}

		path := strings.Split(parsedURL.Path, "/")
		if parsedURL.Host != "news.google.com" || len(path) <= 1 || (path[len(path)-2] != "articles" && path[len(path)-2] != "read") {
			results[i] = DecodeResult{Status: false, Message: "invalid Google News URL"}
			continue
		}

		base64Str := path[len(path)-1]
		decodedBytes, err := base64.URLEncoding.DecodeString(base64Str + "==")
		if err != nil {
			results[i] = DecodeResult{Status: false, Message: fmt.Sprintf("failed to decode base64: %v", err)}
			continue
		}
		decodedStr := string(decodedBytes)

		// Remove prefix
		prefix := string([]byte{0x08, 0x13, 0x22})
		if strings.HasPrefix(decodedStr, prefix) {
			decodedStr = decodedStr[len(prefix):]
		}

		// Remove suffix
		suffix := string([]byte{0xD2, 0x01, 0x00})
		if strings.HasSuffix(decodedStr, suffix) {
			decodedStr = decodedStr[:len(decodedStr)-len(suffix)]
		}

		// Extract URL based on length byte
		if len(decodedStr) > 0 {
			length := int(decodedStr[0])
			if length >= 0x80 && len(decodedStr) > length+1 {
				decodedStr = decodedStr[2 : length+1]
			} else if len(decodedStr) > length+1 {
				decodedStr = decodedStr[1 : length+1]
			}
		}

		// If URL starts with AU_yqL, add to batch
		if strings.HasPrefix(decodedStr, "AU_yqL") {
			batchIDs = append(batchIDs, base64Str)
			idToIndex[base64Str] = i
		} else {
			results[i] = DecodeResult{Status: true, DecodedURL: decodedStr}
		}
	}

	// Process batch IDs
	if len(batchIDs) > 0 {
		batchResult, err := fetchDecodedBatchExecuteMultiple(batchIDs, client)
		if err != nil {
			for _, id := range batchIDs {
				idx := idToIndex[id]
				results[idx] = DecodeResult{Status: false, Message: fmt.Sprintf("batch execute failed: %v", err)}
			}
		} else if batchResult.Status {
			for j, decodedURL := range batchResult.URLs {
				if j < len(batchIDs) {
					idx := idToIndex[batchIDs[j]]
					results[idx] = DecodeResult{Status: true, DecodedURL: decodedURL}
				}
			}
		}
	}

	return results
}

// extractDataAttributes extracts signature and timestamp from Google News HTML page
func extractDataAttributes(htmlContent string) (signature, timestamp string, err error) {
	// Look for data-n-a-sg attribute
	sgRegex := regexp.MustCompile(`data-n-a-sg="([^"]+)"`)
	sgMatch := sgRegex.FindStringSubmatch(htmlContent)
	if sgMatch == nil || len(sgMatch) < 2 {
		return "", "", errors.New("signature not found in HTML")
	}
	signature = sgMatch[1]

	// Look for data-n-a-ts attribute
	tsRegex := regexp.MustCompile(`data-n-a-ts="([^"]+)"`)
	tsMatch := tsRegex.FindStringSubmatch(htmlContent)
	if tsMatch == nil || len(tsMatch) < 2 {
		return "", "", errors.New("timestamp not found in HTML")
	}
	timestamp = tsMatch[1]

	return signature, timestamp, nil
}

// getDecodingParams fetches signature and timestamp required for decoding from Google News
func getDecodingParams(base64Str string, client *http.Client) DecodingParams {
	// Try the articles URL first
	articleURL := fmt.Sprintf("https://news.google.com/articles/%s", base64Str)
	req, err := http.NewRequest("GET", articleURL, nil)
	if err != nil {
		return DecodingParams{Status: false, Message: fmt.Sprintf("failed to create request: %v", err)}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		sig, ts, err := extractDataAttributes(string(body))
		if err == nil {
			return DecodingParams{
				Status:    true,
				Signature: sig,
				Timestamp: ts,
				Base64Str: base64Str,
			}
		}
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Fallback to RSS URL
	rssURL := fmt.Sprintf("https://news.google.com/rss/articles/%s", base64Str)
	req, err = http.NewRequest("GET", rssURL, nil)
	if err != nil {
		return DecodingParams{Status: false, Message: fmt.Sprintf("failed to create RSS request: %v", err)}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	resp, err = client.Do(req)
	if err != nil {
		return DecodingParams{Status: false, Message: fmt.Sprintf("request error: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return DecodingParams{Status: false, Message: fmt.Sprintf("RSS request failed with status: %d", resp.StatusCode)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DecodingParams{Status: false, Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	sig, ts, err := extractDataAttributes(string(body))
	if err != nil {
		return DecodingParams{Status: false, Message: fmt.Sprintf("failed to extract attributes: %v", err)}
	}

	return DecodingParams{
		Status:    true,
		Signature: sig,
		Timestamp: ts,
		Base64Str: base64Str,
	}
}

// decodeURLWithParams decodes the Google News URL using signature and timestamp
func decodeURLWithParams(signature, timestamp, base64Str string, client *http.Client) DecodeResult {
	apiURL := "https://news.google.com/_/DotsSplashUi/data/batchexecute"

	payload := []interface{}{
		"Fbv4je",
		fmt.Sprintf(`["garturlreq",[["X","X",["X","X"],null,null,1,1,"US:en",null,1,null,null,null,null,null,0,1],"X","X",1,[1,1,1],1,1,null,0,0,null,0],"%s",%s,"%s"]`, base64Str, timestamp, signature),
	}

	payloadJSON, err := json.Marshal([][]interface{}{{payload}})
	if err != nil {
		return DecodeResult{Status: false, Message: fmt.Sprintf("failed to marshal payload: %v", err)}
	}

	formData := url.Values{}
	formData.Set("f.req", string(payloadJSON))

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return DecodeResult{Status: false, Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return DecodeResult{Status: false, Message: fmt.Sprintf("request error: %v", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DecodeResult{Status: false, Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	// Parse the response - split by double newline and parse JSON
	parts := strings.SplitN(string(body), "\n\n", 2)
	if len(parts) < 2 {
		return DecodeResult{Status: false, Message: "invalid response format"}
	}

	var parsed []interface{}
	if err := json.Unmarshal([]byte(parts[1]), &parsed); err != nil {
		return DecodeResult{Status: false, Message: fmt.Sprintf("failed to parse response JSON: %v", err)}
	}

	// Navigate the nested structure to get the decoded URL
	if len(parsed) < 1 {
		return DecodeResult{Status: false, Message: "empty response"}
	}

	// The structure is: [[["...",null,"[\"...\",\"decoded_url\"]"]]]
	outerArr, ok := parsed[0].([]interface{})
	if !ok || len(outerArr) < 3 {
		return DecodeResult{Status: false, Message: "unexpected response structure"}
	}

	innerJSON, ok := outerArr[2].(string)
	if !ok {
		return DecodeResult{Status: false, Message: "failed to extract inner JSON"}
	}

	var innerData []interface{}
	if err := json.Unmarshal([]byte(innerJSON), &innerData); err != nil {
		return DecodeResult{Status: false, Message: fmt.Sprintf("failed to parse inner JSON: %v", err)}
	}

	if len(innerData) < 2 {
		return DecodeResult{Status: false, Message: "decoded URL not found in response"}
	}

	decodedURL, ok := innerData[1].(string)
	if !ok {
		return DecodeResult{Status: false, Message: "decoded URL is not a string"}
	}

	return DecodeResult{Status: true, DecodedURL: decodedURL}
}

// NewDecoderV1 decodes Google News URLs using the new method with signature and timestamp.
// This is the recommended decoder for most use cases.
func NewDecoderV1(sourceURL string, interval *time.Duration) DecodeResult {
	client := &http.Client{Timeout: 30 * time.Second}
	return newDecoderV1WithClient(sourceURL, interval, client)
}

func newDecoderV1WithClient(sourceURL string, interval *time.Duration, client *http.Client) DecodeResult {
	// Extract base64 string
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return DecodeResult{Status: false, Message: fmt.Sprintf("failed to parse URL: %v", err)}
	}

	path := strings.Split(parsedURL.Path, "/")
	if parsedURL.Host != "news.google.com" || len(path) <= 1 {
		return DecodeResult{Status: false, Message: "invalid Google News URL format"}
	}

	pathType := path[len(path)-2]
	if pathType != "articles" && pathType != "read" {
		return DecodeResult{Status: false, Message: "invalid Google News URL format"}
	}

	base64Str := path[len(path)-1]

	// Get decoding parameters
	params := getDecodingParams(base64Str, client)
	if !params.Status {
		return DecodeResult{Status: false, Message: params.Message}
	}

	// Decode URL
	result := decodeURLWithParams(params.Signature, params.Timestamp, params.Base64Str, client)

	// Apply interval if specified
	if interval != nil {
		time.Sleep(*interval)
	}

	return result
}
