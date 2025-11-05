package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/sync/errgroup"

	"LegoManagerAPI/internal/config/bricklink"
)

func NewBricklinkService(cfg bricklink.BricklinkConfig) *BricklinkService {
	return &BricklinkService{
		credentials: cfg,
		baseURL:     "https://api.bricklink.com/api/store/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetMinifigComplete fetches all minifig data concurrenlty
func (s *BricklinkService) GetMinifigComplete(ctx context.Context, minifigID string) (*MinifigComplete, error) {
	startTime := time.Now()

	result := &MinifigComplete{
		IndividualFetchTimeMs: make(map[string]int64),
	}

	g, gCtx := errgroup.WithContext(ctx)

	// Fetch info
	g.Go(func() error {
		startInfo := time.Now()
		info, err := s.GetMinifigInfo(gCtx, minifigID)
		result.IndividualFetchTimeMs["info"] = time.Since(startInfo).Milliseconds()
		if err != nil {
			return fmt.Errorf("failed to fetch minifig info: %w", err)
		}
		result.Info = info
		return nil
	})

	// Fetch subsets
	g.Go(func() error {
		startSubsets := time.Now()
		subsets, err := s.GetMinifigSubsets(gCtx, minifigID)
		result.IndividualFetchTimeMs["subsets"] = time.Since(startSubsets).Milliseconds()
		if err != nil {
			return fmt.Errorf("failed to fetch minifig subsets: %w", err)
		}
		result.Subsets = subsets
		return nil
	})

	// Fetch price
	g.Go(func() error {
		startPrice := time.Now()
		price, err := s.GetMinifigPrice(gCtx, minifigID)
		result.IndividualFetchTimeMs["price"] = time.Since(startPrice).Milliseconds()
		if err != nil {
			return fmt.Errorf("failed to fetch minifig price: %w", err)
		}
		result.Price = price
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	result.FetchTimeMs = time.Since(startTime).Milliseconds()

	log.Info("Minifig data fetched",
		"minifig_id", minifigID,
		"total_time_ms", result.FetchTimeMs,
		"info_time_ms", result.IndividualFetchTimeMs["info"],
		"subsets_time_ms", result.IndividualFetchTimeMs["subsets"],
		"price_time_ms", result.IndividualFetchTimeMs["price"])

	return result, nil
}

// GetMinifigInfo fetches minifig basic info
func (s *BricklinkService) GetMinifigInfo(ctx context.Context, minifigID string) (*MinifigInfo, error) {
	endpoint := fmt.Sprintf("/items/MINIFIG/%s", minifigID)

	var resp BricklinkResponse[MinifigInfo]
	if err := s.makeRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// GetMinifigSubsets fetches minifig subsets
func (s *BricklinkService) GetMinifigSubsets(ctx context.Context, minifigID string) (MinifigSubsets, error) {
	endpoint := fmt.Sprintf("/items/MINIFIG/%s/subsets", minifigID)

	var resp BricklinkResponse[MinifigSubsets]
	if err := s.makeRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetMinifigPrice fetches minifig price data
func (s *BricklinkService) GetMinifigPrice(ctx context.Context, minifigID string) (*MinifigPrice, error) {
	endpoint := fmt.Sprintf("/items/MINIFIG/%s/price", minifigID)

	// Price endpoint needs query params
	params := url.Values{}
	params.Set("new_or_used", "N")
	params.Set("currency_code", "USD")

	var resp BricklinkResponse[MinifigPrice]
	if err := s.makeRequest(ctx, "GET", endpoint, params, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// makeRequest handles OAuth1 signing and HTTP request
func (s *BricklinkService) makeRequest(ctx context.Context, method, endpoint string, params url.Values, result interface{}) error {
	fullURL := s.baseURL + endpoint

	// Add OAuth1 parameters
	if params == nil {
		params = url.Values{}
	}

	// Generate OAuth1 signature
	oauthParams := s.generateOAuthParams()
	signedURL, err := s.signRequest(method, fullURL, params, oauthParams)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, signedURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set OAuth header
	req.Header.Set("Authorization", s.buildAuthHeader(oauthParams))
	req.Header.Set("Content-Type", "application/json")

	// perform request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Decode JSON
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// OAuth1 helper functions
func (s *BricklinkService) generateOAuthParams() map[string]string {
	nonce := make([]byte, 16)
	rand.Read(nonce)

	return map[string]string{
		"oauth_consumer_key":     s.credentials.ConsumerKey,
		"oauth_token":            s.credentials.AccessToken,
		"oauth_signature_method": s.credentials.SignatureMethod,
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_nonce":            base64.StdEncoding.EncodeToString(nonce),
		"oauth_version":          "1.0",
	}
}

func (s *BricklinkService) signRequest(method, baseURL string, params url.Values, oauthParams map[string]string) (string, error) {
	// Combine all parameters
	allParams := url.Values{}
	for k, v := range params {
		allParams[k] = v
	}
	for k, v := range oauthParams {
		allParams.Set(k, v)
	}

	// Build signature base string
	encodedParams := s.encodeParameters(allParams)
	signatureBase := fmt.Sprintf("%s&%s&%s",
		url.QueryEscape(method),
		url.QueryEscape(baseURL),
		url.QueryEscape(encodedParams))

	// Create signing key
	signingKey := fmt.Sprintf("%s&%s",
		url.QueryEscape(s.credentials.ConsumerSecret),
		url.QueryEscape(s.credentials.AccessTokenSecret))

	// Generate signature
	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBase))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	oauthParams["oauth_signature"] = signature

	// Build final URL
	if len(params) > 0 {
		return fmt.Sprintf("%s?%s", baseURL, params.Encode()), nil
	}
	return baseURL, nil
}

func (s *BricklinkService) encodeParameters(params url.Values) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(params))
	for _, k := range keys {
		for _, v := range params[k] {
			pairs = append(pairs, fmt.Sprintf("%s=%s",
				url.QueryEscape(k),
				url.QueryEscape(v)))
		}
	}

	return strings.Join(pairs, "&")
}

func (s *BricklinkService) buildAuthHeader(oauthParams map[string]string) string {
	pairs := make([]string, 0, len(oauthParams))
	for k, v := range oauthParams {
		pairs = append(pairs, fmt.Sprintf(`%s="%s"`,
			url.QueryEscape(k),
			url.QueryEscape(v)))
	}
	sort.Strings(pairs)

	return "OAuth " + strings.Join(pairs, ", ")
}
