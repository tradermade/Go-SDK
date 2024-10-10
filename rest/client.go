package tradermade

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://marketdata.tradermade.com/api/v1"

// Structure for the entire API response for live rates
type LiveRate struct {
	Endpoint      string  `json:"endpoint"`
	Quotes        []Quote `json:"quotes"`
	RequestedTime string  `json:"requested_time"`
	Timestamp     int64   `json:"timestamp"`
}

// Structure for individual quotes (for both currency pairs and instruments like indices)
type Quote struct {
	Ask           float64 `json:"ask"`
	Bid           float64 `json:"bid"`
	Mid           float64 `json:"mid"`
	BaseCurrency  string  `json:"base_currency,omitempty"`  // Optional field
	QuoteCurrency string  `json:"quote_currency,omitempty"` // Optional field
	Instrument    string  `json:"instrument,omitempty"`     // Optional field for indices
}
type HistoricalRate struct {
	Date        string            `json:"date"`
	Endpoint    string            `json:"endpoint"`
	Quotes      []HistoricalQuote `json:"quotes"`
	RequestTime string            `json:"request_time"`
}
type ConvertResponse struct {
	BaseCurrency  string  `json:"base_currency"`
	QuoteCurrency string  `json:"quote_currency"`
	Quote         float64 `json:"quote"`
	Total         float64 `json:"total"`
	RequestedTime string  `json:"requested_time"`
	Timestamp     int64   `json:"timestamp"`
}

// Structure for handling API error responses
type ErrorResponseOK struct {
	Error   int    `json:"error"`   // Numeric error code for 200 responses
	Message string `json:"message"` // Error message
}

type ErrorResponse struct {
	Message string                 `json:"message"` // The general error message
	Errors  map[string]interface{} `json:"errors"`  // The specific error messages in a key-value map

}

// HistoricalQuote represents an individual quote in the historical response
type HistoricalQuote struct {
	BaseCurrency  string  `json:"base_currency"`
	QuoteCurrency string  `json:"quote_currency"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
}

type HistoricalData struct {
	Endpoint    string  `json:"endpoint"`
	Currency    string  `json:"currency"`
	DateTime    string  `json:"date_time"`
	Open        float64 `json:"open"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	Close       float64 `json:"close"`
	RequestTime string  `json:"request_time"`
}

// Structure for parsing timeseries data
type TimeSeriesRate struct {
	BaseCurrency  string            `json:"base_currency"`
	QuoteCurrency string            `json:"quote_currency"`
	StartDate     string            `json:"start_date"`
	EndDate       string            `json:"end_date"`
	Endpoint      string            `json:"endpoint"`
	Quotes        []TimeSeriesQuote `json:"quotes"`
	RequestTime   string            `json:"request_time"`
}

// Structure for individual quotes in the timeseries response
type TimeSeriesQuote struct {
	Date  string  `json:"date"`
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
}

// RESTClient structure that includes the HTTP client and API key
type RESTClient struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewRESTClient initializes a new REST client
func NewRESTClient(apiKey string) *RESTClient {
	return &RESTClient{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// GetLiveRates fetches live rates for specified currencies or instruments
func (c *RESTClient) GetLiveRates(currencies []string) (*LiveRate, error) {
	// Construct the URL
	URL := fmt.Sprintf("https://marketdata.tradermade.com/api/v1/live?currency=%s&api_key=%s", joinStrings(currencies), c.APIKey)

	encodedURL := strings.ReplaceAll(URL, " ", "%20")
	resp, err := c.HTTPClient.Get(encodedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		// Try to decode the error response
		var errorResponse ErrorResponse

		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API request failed with status code %d: %v ", resp.StatusCode, errorResponse.Errors)
	}

	// DEBUG: Print the raw response body to check the content
	//	fmt.Printf("Raw Response Body: %s\n", string(body))

	// Check if the JSON response contains an "error" field
	var errorResponse ErrorResponseOK
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		// If the error field is not empty, return it as an error
		if errorResponse.Error != 0 {
			return nil, fmt.Errorf("API error: %d - %s", errorResponse.Error, errorResponse.Message)
		}
	}
	// Check if the status code is not OK
	// Check if the response contains an error message even with a 200 status code

	var liveRate LiveRate
	if err := json.Unmarshal(body, &liveRate); err != nil {
		return nil, fmt.Errorf("failed to parse successful response: %v", err)
	}

	return &liveRate, nil
}

func (c *RESTClient) GetHistoricalRates(currency, dateTime, interval string) (interface{}, error) {
	var URL string
	switch interval {
	case "minute":
		URL = fmt.Sprintf("%s/minute_historical?currency=%s&date_time=%s&api_key=%s", baseURL, currency, dateTime, c.APIKey)
		var minuteRate HistoricalData
		if err := c.sendHistoricalRequest(URL, &minuteRate); err != nil {
			return nil, err
		}
		return &minuteRate, nil
	case "hour":
		URL = fmt.Sprintf("%s/hour_historical?currency=%s&date_time=%s&api_key=%s", baseURL, currency, dateTime, c.APIKey)
		var hourRate HistoricalData
		if err := c.sendHistoricalRequest(URL, &hourRate); err != nil {
			return nil, err
		}
		return &hourRate, nil
	case "day":
		URL = fmt.Sprintf("%s/historical?currency=%s&date=%s&api_key=%s", baseURL, currency, dateTime, c.APIKey)
		var dailyRate HistoricalRate
		if err := c.sendHistoricalRequest(URL, &dailyRate); err != nil {
			return nil, err
		}
		return &dailyRate, nil
	default:
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}
}

// GetTimeSeriesData fetches time series data for a given currency and date range
func (c *RESTClient) GetTimeSeriesData(
	currency string,
	startDate string,
	endDate string,
	interval string, // "daily", "hourly", or "minute"
	period ...int) (*TimeSeriesRate, error) {

	// Validate and construct URL based on interval
	var URL string

	// Base URL for timeseries endpoint with mandatory fields
	baseURL := fmt.Sprintf("%s/timeseries?currency=%s&start_date=%s&end_date=%s&format=records&api_key=%s",
		baseURL, currency, startDate, endDate, c.APIKey)

	// If interval is daily, no period is required
	if strings.ToLower(interval) == "daily" {
		URL = baseURL + "&interval=daily"
	} else if strings.ToLower(interval) == "hourly" || strings.ToLower(interval) == "minute" {
		// Check if the period is provided for hourly or minute intervals
		if len(period) == 0 {
			return nil, fmt.Errorf("period must be provided for %s interval", interval)
		}

		// Handle hourly interval with period validation
		if strings.ToLower(interval) == "hourly" {
			if isValidPeriodForHourly(period[0]) {
				URL = fmt.Sprintf("%s&interval=hourly&period=%d", baseURL, period[0])
			} else {
				return nil, fmt.Errorf("invalid period for hourly interval: %d", period[0])
			}
		}

		// Handle minute interval with period validation
		if strings.ToLower(interval) == "minute" {
			if isValidPeriodForMinute(period[0]) {
				URL = fmt.Sprintf("%s&interval=minute&period=%d", baseURL, period[0])
			} else {
				return nil, fmt.Errorf("invalid period for minute interval: %d", period[0])
			}
		}
	} else {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	// encode url to eliminate space
	encodedURL := strings.ReplaceAll(URL, " ", "%20")
	resp, err := c.HTTPClient.Get(encodedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if the status code is not OK
	if resp.StatusCode != http.StatusOK {
		// Try to decode the error response
		var errorResponse ErrorResponse

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API request failed with status code %d: %v ", resp.StatusCode, formatErrorMap(errorResponse.Errors))
	}

	var errorResponse ErrorResponseOK
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		// If the error field is not empty, return it as an error
		if errorResponse.Error != 0 {
			return nil, fmt.Errorf("API error: %d - %s", errorResponse.Error, errorResponse.Message)
		}
	}

	// Decode the successful response into the TimeSeriesRate struct
	var timeSeriesData TimeSeriesRate
	if err := json.Unmarshal(body, &timeSeriesData); err != nil {
		return nil, fmt.Errorf("failed to parse successful response: %v", err)
	}

	return &timeSeriesData, nil
}

// ConvertCurrency sends a request to the TraderMade Convert API
func (c *RESTClient) ConvertCurrency(from string, to string, amount float64) (*ConvertResponse, error) {
	// Construct the URL
	URL := fmt.Sprintf("https://marketdata.tradermade.com/api/v1/convert?from=%s&to=%s&amount=%f&api_key=%s",
		from, to, amount, c.APIKey)

	// Perform the request
	encodedURL := strings.ReplaceAll(URL, " ", "")
	resp, err := c.HTTPClient.Get(encodedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to decode the error response
		var errorResponse ErrorResponse

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API request failed with status code %d: %v", resp.StatusCode, errorResponse.Errors)
	}

	// Check if the status code is not OK
	var errorResponse ErrorResponseOK
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		// If the error field is not empty, return it as an error
		if errorResponse.Error != 0 {
			return nil, fmt.Errorf("API error: %d - %s", errorResponse.Error, errorResponse.Message)
		}
	}

	// Decode the successful response into the ConvertResponse struct
	var convertResponse ConvertResponse
	if err := json.Unmarshal(body, &convertResponse); err != nil {
		return nil, fmt.Errorf("failed to parse successful response: %v", err)
	}

	return &convertResponse, nil
}

// sendHistoricalRequest is a helper function to make the HTTP request and unmarshal the response
func (c *RESTClient) sendHistoricalRequest(URL string, v interface{}) error {
	encodedURL := strings.ReplaceAll(URL, " ", "%20")
	resp, err := c.HTTPClient.Get(encodedURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if the status code is not OK
	if resp.StatusCode != http.StatusOK {
		// Try to decode the error response
		var errorResponse ErrorResponse

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))

		}
		return fmt.Errorf("not nil API request failed with status code %d: %v", resp.StatusCode, errorResponse.Errors)

	}
	var errorResponse ErrorResponseOK
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		// If the error field is not empty, return it as an error
		if errorResponse.Error != 0 {
			return fmt.Errorf("API error: %d - %s", errorResponse.Error, errorResponse.Message)
		}
	}

	// Decode the successful response into the provided interface (v)
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to parse successful response: %v", err)
	}

	return nil
}

// Helper function to join currency pairs into a single string
func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}

// isValidPeriodForHourly validates the period for hourly intervals
func isValidPeriodForHourly(period int) bool {
	validPeriods := []int{1, 2, 4, 6, 8, 24}
	for _, p := range validPeriods {
		if period == p {
			return true
		}
	}
	return false
}

// isValidPeriodForMinute validates the period for minute intervals
func isValidPeriodForMinute(period int) bool {
	validPeriods := []int{1, 5, 10, 15, 30}
	for _, p := range validPeriods {
		if period == p {
			return true
		}
	}
	return false
}
func formatErrorMap(errors map[string]interface{}) string {
	var formattedErrors string
	for key, value := range errors {
		// Safely convert value to string for readability
		formattedErrors += fmt.Sprintf("%s: %v; ", key, value)
	}
	return formattedErrors
}
