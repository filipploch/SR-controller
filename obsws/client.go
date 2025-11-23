package obsws

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client reprezentuje klienta OBS-WebSocket
type Client struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	callbacks map[string]chan map[string]interface{}
	requestID int
	address   string
	reconnect bool
	connected bool
}

// Message reprezentuje wiadomość OBS-WebSocket
type Message struct {
	Op int                    `json:"op"`
	D  map[string]interface{} `json:"d"`
}

// NewClient tworzy nowego klienta OBS-WebSocket
func NewClient(address string) (*Client, error) {
	client := &Client{
		callbacks: make(map[string]chan map[string]interface{}),
		requestID: 1,
		address:   address,
		reconnect: true,
		connected: false,
	}

	if err := client.connect(); err != nil {
		return nil, err
	}

	return client, nil
}

// connect nawiązuje połączenie z OBS
func (c *Client) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.address, nil)
	if err != nil {
		return fmt.Errorf("błąd połączenia z OBS-WebSocket: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.mu.Unlock()

	// Uruchom goroutine do odbierania wiadomości
	go c.receiveMessages()

	// Identyfikuj się (op code 1)
	identifyMsg := Message{
		Op: 1,
		D: map[string]interface{}{
			"rpcVersion": 1,
		},
	}

	if err := c.send(identifyMsg); err != nil {
		return err
	}

	log.Println("Połączono z OBS-WebSocket")
	return nil
}

// send wysyła wiadomość do OBS
func (c *Client) send(msg Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(msg)
}

// receiveMessages odbiera wiadomości z OBS
func (c *Client) receiveMessages() {
	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Błąd odczytu z OBS-WebSocket: %v", err)

			c.mu.Lock()
			c.connected = false
			c.mu.Unlock()

			// Jeśli reconnect jest włączony, próbuj ponownie
			if c.reconnect {
				log.Println("Próba ponownego połączenia z OBS...")
				for {
					time.Sleep(5 * time.Second)
					if err := c.connect(); err != nil {
						log.Printf("Nie można połączyć: %v, ponowna próba za 5s...", err)
						continue
					}
					log.Println("Pomyślnie połączono ponownie z OBS")
					return // Nowy goroutine został uruchomiony w connect()
				}
			}
			return
		}

		// Obsługa odpowiedzi na żądania (op code 7)
		if msg.Op == 7 {
			if requestID, ok := msg.D["requestId"].(string); ok {
				c.mu.Lock()
				if ch, exists := c.callbacks[requestID]; exists {
					ch <- msg.D
					delete(c.callbacks, requestID)
				}
				c.mu.Unlock()
			}
		}
	}
}

// Request wysyła żądanie do OBS i czeka na odpowiedź
func (c *Client) Request(requestType string, requestData map[string]interface{}) (map[string]interface{}, error) {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return nil, fmt.Errorf("brak połączenia z OBS")
	}

	requestID := fmt.Sprintf("req-%d", c.requestID)
	c.requestID++

	responseChan := make(chan map[string]interface{}, 1)
	c.callbacks[requestID] = responseChan
	c.mu.Unlock()

	requestMsg := Message{
		Op: 6, // Request op code
		D: map[string]interface{}{
			"requestType": requestType,
			"requestId":   requestID,
		},
	}

	if requestData != nil {
		requestMsg.D["requestData"] = requestData
	}

	if err := c.send(requestMsg); err != nil {
		c.mu.Lock()
		delete(c.callbacks, requestID)
		c.mu.Unlock()
		return nil, err
	}

	// Czekaj na odpowiedź
	response := <-responseChan

	// Sprawdź status
	if status, ok := response["requestStatus"].(map[string]interface{}); ok {
		if result, ok := status["result"].(bool); ok && !result {
			if comment, ok := status["comment"].(string); ok {
				return response, fmt.Errorf("żądanie nieudane: %s", comment)
			}
			return response, fmt.Errorf("żądanie nieudane")
		}
	}

	return response, nil
}

// SetSourceVisibility ustawia widoczność źródła w scenie
func (c *Client) SetSourceVisibility(sceneName, sourceName string, visible bool) error {
	_, err := c.Request("SetSceneItemEnabled", map[string]interface{}{
		"sceneName":        sceneName,
		"sceneItemId":      c.getSceneItemID(sceneName, sourceName),
		"sceneItemEnabled": visible,
	})
	return err
}

// SetSceneItemIndex ustawia pozycję źródła w scenie (0 = najwyżej)
func (c *Client) SetSceneItemIndex(sceneName, sourceName string, toTop bool) error {
	sceneItemID := c.getSceneItemID(sceneName, sourceName)
	if sceneItemID == 0 {
		return fmt.Errorf("nie znaleziono źródła %s w scenie %s", sourceName, sceneName)
	}

	// Jeśli chcemy na górę, musimy pobrać liczbę źródeł
	if toTop {
		items, err := c.GetSceneItemList(sceneName)
		if err != nil {
			return err
		}
		// Największy indeks = górna pozycja
		topIndex := len(items) - 1

		_, err = c.Request("SetSceneItemIndex", map[string]interface{}{
			"sceneName":      sceneName,
			"sceneItemId":    sceneItemID,
			"sceneItemIndex": topIndex,
		})
		return err
	}

	// Index 0 = dół
	_, err := c.Request("SetSceneItemIndex", map[string]interface{}{
		"sceneName":      sceneName,
		"sceneItemId":    sceneItemID,
		"sceneItemIndex": 0,
	})
	return err
}

// SetCurrentProgramScene ustawia aktywną scenę (program scene)
func (c *Client) SetCurrentProgramScene(sceneName string) error {
	_, err := c.Request("SetCurrentProgramScene", map[string]interface{}{
		"sceneName": sceneName,
	})
	return err
}

// getSceneItemID pobiera ID elementu sceny (uproszczona wersja - wymaga rozbudowy)
func (c *Client) getSceneItemID(sceneName, sourceName string) int {
	response, err := c.Request("GetSceneItemId", map[string]interface{}{
		"sceneName":  sceneName,
		"sourceName": sourceName,
	})
	if err != nil {
		log.Printf("Błąd pobierania ID elementu sceny: %v", err)
		return 0
	}

	if responseData, ok := response["responseData"].(map[string]interface{}); ok {
		if itemID, ok := responseData["sceneItemId"].(float64); ok {
			return int(itemID)
		}
	}
	return 0
}

// GetSceneItemList pobiera listę źródeł w scenie
func (c *Client) GetSceneItemList(sceneName string) ([]map[string]interface{}, error) {
	response, err := c.Request("GetSceneItemList", map[string]interface{}{
		"sceneName": sceneName,
	})
	if err != nil {
		return nil, err
	}

	if responseData, ok := response["responseData"].(map[string]interface{}); ok {
		if sceneItems, ok := responseData["sceneItems"].([]interface{}); ok {
			items := make([]map[string]interface{}, len(sceneItems))
			for i, item := range sceneItems {
				if itemMap, ok := item.(map[string]interface{}); ok {
					items[i] = itemMap
				}
			}
			return items, nil
		}
	}

	return nil, fmt.Errorf("nie można pobrać listy źródeł")
}

// SetSceneItemIndexByValue ustawia konkretny indeks dla źródła
func (c *Client) SetSceneItemIndexByValue(sceneName, sourceName string, index int) error {
	sceneItemID := c.getSceneItemID(sceneName, sourceName)
	if sceneItemID == 0 {
		return fmt.Errorf("nie znaleziono źródła: %s w scenie: %s", sourceName, sceneName)
	}

	_, err := c.Request("SetSceneItemIndex", map[string]interface{}{
		"sceneName":      sceneName,
		"sceneItemId":    sceneItemID,
		"sceneItemIndex": index,
	})

	return err
}

// Close zamyka połączenie
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reconnect = false // Wyłącz automatyczne reconnect
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected sprawdza czy klient jest połączony
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}
