# TraderMade Go SDK

This Go SDK provides easy access to TraderMade's forex data API. It allows you to fetch live rates, historical data, perform currency conversions, and retrieve time series data.

## Installation

To use this SDK in your Go project, run:

```bash
go get github.com/tradermade/Go-SDK
```

## Usage

Here's how to use the main features of the TraderMade Go SDK:

### Initializing the Client

```go
import (
    tradermade "github.com/tradermade/Go-SDK/rest"
)

client := tradermade.NewRESTClient("YOUR_API_KEY")
```

### Fetching Live Rates

```go
currencyPairs := []string{"EURUSD", "GBPUSD", "USDJPY"}
liveRates, err := client.GetLiveRates(currencyPairs)
if err != nil {
    log.Fatalf("Error fetching live rates: %v", err)
}

for _, quote := range liveRates.Quotes {
    fmt.Printf("Base: %s, Quote: %s, Bid: %f, Ask: %f, Mid: %f\n",
        quote.BaseCurrency, quote.QuoteCurrency, quote.Bid, quote.Ask, quote.Mid)
}
```

### Fetching Historical Rates

#### Daily Data

```go
currency := "EURUSD"
date := time.Now().AddDate(0, 0, -1).Format("2006-01-02") // yesterday's date
interval := "day"

historicalRates, err := client.GetHistoricalRates(currency, date, interval)
if err != nil {
    log.Fatalf("Error fetching daily historical rates: %v", err)
}

if rates, ok := historicalRates.(*tradermade.HistoricalRate); ok {
    for _, quote := range rates.Quotes {
        fmt.Printf("Date: %s, Open: %f, High: %f, Low: %f, Close: %f\n",
            date, quote.Open, quote.High, quote.Low, quote.Close)
    }
}
```

#### Hourly Data

```go
interval = "hour"
dateTime := time.Now().AddDate(0, 0, -1).Format("2006-01-02-15:00")

hourlyRates, err := client.GetHistoricalRates(currency, dateTime, interval)
if err != nil {
    log.Fatalf("Error fetching hourly historical rates: %v", err)
}

if hourly, ok := hourlyRates.(*tradermade.HistoricalData); ok {
    fmt.Printf("DateTime: %s, Open: %f, High: %f, Low: %f, Close: %f\n",
        hourly.DateTime, hourly.Open, hourly.High, hourly.Low, hourly.Close)
}
```

### Currency Conversion

```go
convertResult, err := client.ConvertCurrency("EUR", "GBP", 1000.0)
if err != nil {
    log.Fatalf("Error fetching conversion data: %v", err)
}

fmt.Printf("Converted %s to %s:\n", convertResult.BaseCurrency, convertResult.QuoteCurrency)
fmt.Printf("Quote: %f\n", convertResult.Quote)
fmt.Printf("Total: %f\n", convertResult.Total)
fmt.Printf("Requested Time: %s\n", convertResult.RequestedTime)
fmt.Printf("Timestamp: %d\n", convertResult.Timestamp)
```

### Time Series Data

#### Daily Data

```go
timeSeriesData, err := client.GetTimeSeriesData("EURUSD", "2019-10-01", "2019-10-10", "daily")
if err != nil {
    log.Fatalf("Error fetching daily time series data: %v", err)
}

fmt.Printf("Time Series Data from %s to %s:\n", timeSeriesData.StartDate, timeSeriesData.EndDate)
for _, quote := range timeSeriesData.Quotes {
    fmt.Printf("Date: %s, Open: %f, High: %f, Low: %f, Close: %f\n",
        quote.Date, quote.Open, quote.High, quote.Low, quote.Close)
}
```

#### Hourly Data

```go
timeSeriesDataHourly, err := client.GetTimeSeriesData("EURUSD", "2024-10-01 10:00", "2024-10-02-11:00", "hourly", 4)
if err != nil {
    log.Fatalf("Error fetching hourly time series data: %v", err)
}

fmt.Printf("Time Series Data (Hourly) from %s to %s:\n",
    timeSeriesDataHourly.StartDate, timeSeriesDataHourly.EndDate)
for _, quote := range timeSeriesDataHourly.Quotes {
    fmt.Printf("Date: %s, Open: %f, High: %f, Low: %f, Close: %f\n",
        quote.Date, quote.Open, quote.High, quote.Low, quote.Close)
}
```

#### Minute Data

```go
timeSeriesDataMinute, err := client.GetTimeSeriesData("EURUSD", "2024-10-02", "2024-10-02-23:59", "minute", 15)
if err != nil {
    log.Fatalf("Error fetching minute time series data: %v", err)
}

fmt.Printf("Time Series Data (Minute) from %s to %s:\n",
    timeSeriesDataMinute.StartDate, timeSeriesDataMinute.EndDate)
for _, quote := range timeSeriesDataMinute.Quotes {
    fmt.Printf("Date: %s, Open: %f, High: %f, Low: %f, Close: %f\n",
        quote.Date, quote.Open, quote.High, quote.Low, quote.Close)
}
```

## Error Handling

All methods return an error as the second return value. Always check this error before using the returned data.

## API Documentation

For more details on the TraderMade API, please refer to the [official API documentation](https://tradermade.com/docs/api-documentation).

## Support

If you encounter any issues or have questions, please open an issue on the [GitHub repository](https://github.com/tradermade/Go-SDK) or contact TraderMade support.

## License

[Include your license information here]
