package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tradermadews "github.com/tradermade/Go-SDK/websocket"
)

func main() {
	// Initialize the WebSocket client with your API key
	client := tradermadews.NewWebSocketClient("add_ws_key", "EURUSD,GBPUSD,XAUUSD")

	// Set a handler for the "Connected" message
	client.SetConnectedHandler(func(connectedMsg tradermadews.ConnectedMessage) {
		fmt.Printf("WebSocket connected: %s\n", connectedMsg.Message)
	})

	// Set a message handler to print the entire raw message
	client.SetMessageHandler(func(quote tradermadews.QuoteMessage, humanTimestamp string) {
		fmt.Printf("Received quote: Symbol=%s Bid=%.5f Ask=%.5f Timestamp=%s (Human-readable: %s)\n",
			quote.Symbol, quote.Bid, quote.Ask, quote.Ts, humanTimestamp)
	})

	// Set a handler to notify reconnection attempts
	client.SetReconnectionHandler(func(attempt int) {
		fmt.Printf("Reconnecting... (Attempt %d)\n", attempt)
	})

	// Set custom retry settings
	client.MaxRetries = 10                 // Set
	client.RetryInterval = 5 * time.Second // Set retry interval

	// Enable automatic reconnection
	client.EnableAutoReconnect(true)

	// Connect to the TraderMade WebSocket
	err := client.Connect()
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
	}

	// Handle graceful shutdown (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	// Disconnect WebSocket on shutdown
	client.Disconnect()

	fmt.Println("Shutting down WebSocket client...")
}
