package tradermadews

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const wsURL = "wss://marketdata.tradermade.com/feedadv"

// QuoteMessage represents a quote from the WebSocket feed
type QuoteMessage struct {
	Symbol string  `json:"symbol"`
	Bid    float64 `json:"bid"`
	Ask    float64 `json:"ask"`
	Mid    float64 `json:"mid"`
	Ts     string  `json:"ts"` // Timestamp as a string (from API response)
}

// ConnectedMessage represents the connection status message
type ConnectedMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// WebSocketClient manages the WebSocket connection and state
type WebSocketClient struct {
	APIKey              string
	Symbol              string // Single string for the symbol to subscribe to
	Conn                *websocket.Conn
	ConnMutex           sync.Mutex
	MessageHandler      func(QuoteMessage, string) // Handles market data with a human-readable timestamp
	ConnectedHandler    func(ConnectedMessage)     // Handles the "Connected" message
	ReconnectionHandler func(int)                  // Handles reconnection attempts

	MaxRetries    int           // Maximum retries for reconnection
	RetryInterval time.Duration // Time between retries
	AutoReconnect bool          // Enable/Disable automatic reconnection
	StopReconnect chan struct{} // Channel to stop reconnection attempts
}

// NewWebSocketClient initializes the WebSocket client with an API key and symbol
func NewWebSocketClient(apiKey, symbol string) *WebSocketClient {
	return &WebSocketClient{
		APIKey:        apiKey,
		Symbol:        symbol,
		MaxRetries:    5,               // Default maximum retries
		RetryInterval: 5 * time.Second, // Default retry interval
		AutoReconnect: true,            // Auto-reconnect enabled by default
		StopReconnect: make(chan struct{}),
	}
}

// SetSymbol sets the symbol for WebSocket streaming
func (client *WebSocketClient) SetSymbol(symbol string) {
	client.Symbol = symbol
}

// SetMessageHandler sets the callback function to handle incoming WebSocket messages
func (client *WebSocketClient) SetMessageHandler(handler func(QuoteMessage, string)) {
	client.MessageHandler = handler
}

// SetConnectedHandler sets the callback function to handle the "Connected" message
func (client *WebSocketClient) SetConnectedHandler(handler func(ConnectedMessage)) {
	client.ConnectedHandler = handler
}

// SetReconnectionHandler sets the callback function for reconnection attempts
func (client *WebSocketClient) SetReconnectionHandler(handler func(int)) {
	client.ReconnectionHandler = handler
}

// EnableAutoReconnect enables/disables automatic reconnection
func (client *WebSocketClient) EnableAutoReconnect(enable bool) {
	client.AutoReconnect = enable
}

// Connect establishes a WebSocket connection to the TraderMade API
func (client *WebSocketClient) Connect() error {
	client.ConnMutex.Lock()
	defer client.ConnMutex.Unlock()

	// If connection already exists, don't reconnect
	if client.Conn != nil {
		return nil
	}

	// Establish connection
	var err error
	client.Conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Printf("WebSocket connection failed: %v\n", err)
		return err
	}

	// Start reading messages
	go client.wsReadPump()

	// Send authentication message with user key and symbol
	cred := fmt.Sprintf(`{"userKey":"%s", "symbol":"%s"}`, client.APIKey, client.Symbol)
	err = client.Conn.WriteMessage(websocket.TextMessage, []byte(cred))
	if err != nil {
		return fmt.Errorf("Failed to send credentials: %w", err)
	}

	return nil
}

// Disconnect closes the WebSocket connection and stops reconnection attempts
func (client *WebSocketClient) Disconnect() error {
	close(client.StopReconnect) // Stop reconnect attempts

	client.ConnMutex.Lock()
	defer client.ConnMutex.Unlock()

	if client.Conn != nil {
		err := client.Conn.Close()
		client.Conn = nil
		return err
	}
	return nil
}

// wsReadPump handles incoming messages from the WebSocket connection
func (client *WebSocketClient) wsReadPump() {
	defer func() {
		client.ConnMutex.Lock()
		client.Conn.Close()
		client.Conn = nil
		client.ConnMutex.Unlock()

		if client.AutoReconnect {
			client.reconnect() // Try to reconnect when the connection is closed
		}
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			fmt.Printf("WebSocket read error: %v\n", err)
			return
		}

		// Check if the message is valid JSON (starts with '{' or '[')
		msgStr := string(message)
		if strings.HasPrefix(msgStr, "{") || strings.HasPrefix(msgStr, "[") {
			// Try to handle the "Connected" message
			var connectedMsg ConnectedMessage
			if err := json.Unmarshal(message, &connectedMsg); err == nil && connectedMsg.Status == "connected" {
				if client.ConnectedHandler != nil {
					client.ConnectedHandler(connectedMsg) // Pass "Connected" message to the handler
				}
				continue
			}

			// Parse the JSON message into the QuoteMessage struct (for market data)
			var quote QuoteMessage
			err = json.Unmarshal(message, &quote)
			if err != nil {
				fmt.Printf("Failed to unmarshal quote message: %v\n", err)
				continue
			}

			// Convert the timestamp from string to int64
			tsInt, err := strconv.ParseInt(quote.Ts, 10, 64)
			if err != nil {
				fmt.Printf("Failed to parse timestamp: %v\n", err)
				continue
			}

			// Convert the timestamp from milliseconds to human-readable format (including milliseconds)
			timestamp := time.Unix(0, tsInt*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000")

			// If the handler is set, call it with the parsed quote message and human-readable timestamp
			if client.MessageHandler != nil {
				client.MessageHandler(quote, timestamp)
			}
		} else {
			// Non-JSON message: Handle appropriately (e.g., skip, log, etc.)
			fmt.Printf("Status: %s\n", msgStr)
		}
	}
}

// reconnect attempts to reconnect to the WebSocket with retry logic
func (client *WebSocketClient) reconnect() {
	retries := 0
	for {
		retries++
		if retries > client.MaxRetries {
			fmt.Println("Max retries reached. Stopping reconnection attempts.")
			return
		}

		// Notify reconnection attempt
		if client.ReconnectionHandler != nil {
			client.ReconnectionHandler(retries)
		}

		fmt.Printf("Attempting to reconnect... (Attempt %d/%d)\n", retries, client.MaxRetries)
		err := client.Connect()
		if err == nil {
			fmt.Println("Successfully reconnected to WebSocket.")
			return
		}

		// Wait for the retry interval or stop if requested
		select {
		case <-time.After(client.RetryInterval):
		case <-client.StopReconnect:
			fmt.Println("Reconnect stopped.")
			return
		}
	}
}
