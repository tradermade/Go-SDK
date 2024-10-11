package main

import (
	"fmt"
	"log"
	"time"

	tradermade "github.com/tradermade/Go-SDK/rest"
)

func main() {
	// Initialize the REST client with your API key

	client := tradermade.NewRESTClient("api_key")

	//Fetch live rates for the specified currency pairs
	currencyPairs := []string{"EURUSD", "GBPUSD", "USDJPY"}
	liveRates, err := client.GetLiveRates(currencyPairs)
	if err != nil {
		log.Fatalf("Error fetching live rates: %v", err)
	}

	fmt.Println("Live Rates:")
	for _, quote := range liveRates.Quotes {
		fmt.Printf("Base: %s, Quote: %s, Bid: %f, Ask: %f, Mid: %f\n",
			quote.BaseCurrency, quote.QuoteCurrency, quote.Bid, quote.Ask, quote.Mid)
	}

	// Fetch daily historical rates for EURUSD for a specific date (previous day)
	currency := "EURUSD"
	date := time.Now().AddDate(0, 0, -1).Format("2006-01-02") // yesterday's date
	interval := "day"                                         // daily data

	historicalRates, err := client.GetHistoricalRates(currency, date, interval)
	if err != nil {
		log.Fatalf("Error fetching daily historical rates: %v", err)
	}

	fmt.Println("Historical Rates for the day:")
	if rates, ok := historicalRates.(*tradermade.HistoricalRate); ok {
		for _, quote := range rates.Quotes {
			fmt.Printf("Date: %s, Open: %f, High: %f, Low: %f, Close: %f\n",
				date, quote.Open, quote.High, quote.Low, quote.Close)
		}
	} else {
		fmt.Println("Unexpected response format")
	}

	// Fetch hourly historical rates for EURUSD for the current day
	interval = "hour"
	dateTime := time.Now().AddDate(0, 0, -1).Format("2006-01-02-15:00") // specify the time for hourly data

	hourlyRates, err := client.GetHistoricalRates(currency, dateTime, interval)
	if err != nil {
		log.Fatalf("Error fetching hourly historical rates: %v", err)
	}

	fmt.Println("Historical Hourly Rates:")
	if hourly, ok := hourlyRates.(*tradermade.HistoricalData); ok {
		fmt.Printf("DateTime: %s, Open: %f, High: %f, Low: %f, Close: %f\n",
			hourly.DateTime, hourly.Open, hourly.High, hourly.Low, hourly.Close)
	} else {
		fmt.Println("Unexpected response format for hourly data")
	}
	convertResult, err := client.ConvertCurrency("EUR ", "GBP", 1000.0)
	if err != nil {
		log.Fatalf("Error fetching conversion data: %v", err)
	}

	// Print the conversion result
	fmt.Printf("Converted %s to %s:\n", convertResult.BaseCurrency, convertResult.QuoteCurrency)
	fmt.Printf("Quote: %f\n", convertResult.Quote)
	fmt.Printf("Total: %f\n", convertResult.Total)
	fmt.Printf("Requested Time: %s\n", convertResult.RequestedTime)
	fmt.Printf("Timestamp: %d\n", convertResult.Timestamp)

	timeSeriesData, err := client.GetTimeSeriesData("EURUSD", "2019-10-01", "2019-10-10", "daily")
	if err != nil {
		log.Fatalf("Error fetching daily time series data: %v", err)
	}

	fmt.Printf("Time Series Data from %s to %s:\n", timeSeriesData.StartDate, timeSeriesData.EndDate)
	for _, quote := range timeSeriesData.Quotes {
		fmt.Printf("Date: %s, Open: %f, High: %f, Low: %f, Close: %f\n",
			quote.Date, quote.Open, quote.High, quote.Low, quote.Close)
	}

	//Example for hourly data (period required)
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
	// Example for minute data (period required)
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

}
