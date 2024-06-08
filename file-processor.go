package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type details struct {
	Lines  int `json:"lines"`
	Words  int `json:"words"`
	Letter int `json:"letters"`
}

type message struct {
	Message string `json:"message"`
}

var ctx = context.Background()
var redisClient *redis.Client
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Event bus to handle messages
var eventBus = make(chan string)

func processFile(filename string) (details, error) {
	fileLookUpPath := os.Getenv("FILE_LOOKUP_PATH")
	if fileLookUpPath == "" {
		fileLookUpPath = "/tmp/upload"
	}
	file, err := os.Open(fileLookUpPath + "/" + filename)
	if err != nil {
		return details{}, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines, words, letters := 0, 0, 0
	for scanner.Scan() {
		lines++
		words += len(strings.Fields(scanner.Text()))
		letters += len(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return details{}, fmt.Errorf("error reading file: %v", err)
	}

	result := details{Lines: lines, Words: words, Letter: letters}
	return result, nil
}

func getMessage(c *gin.Context) {
	c.JSON(http.StatusOK, message{Message: "PONG"})
}

// Redis subscribe function
func redisSubscribe(channel string) {
	pubsub := redisClient.Subscribe(ctx, channel)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			fmt.Println("Error subscribing:", err)
			close(eventBus)
			return
		}
		fmt.Println("Message received:", msg.Payload)
		eventBus <- "Got the file Information, Processing It..."
		//eventBus <- msg.Payload
		result, err := processFile(msg.Payload)
		if err != nil {
			eventBus <- fmt.Sprintf("Error processing file: %v", err)
			return
		}
		//eventBus <- fmt.Sprintf("Lines: %d, Words: %d, Letters: %d", result.Lines, result.Words, result.Letter)
		//send json as string
		eventBus <- fmt.Sprintf("{\"Lines\": %d, \"Words\": %d, \"Letters\": %d}", result.Lines, result.Words, result.Letter)
	}
}

func websocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade:", err)
		return
	}
	defer conn.Close()

	// Channel to handle read messages
	clientMessages := make(chan string)

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error reading message:", err)
				close(clientMessages)
				return
			}
			fmt.Println("Message received:", string(msg))
			// Add message to the event bus
			eventBus <- string(msg)
		}
	}()

	// Listen for messages from the event bus and send them to the WebSocket client
	for {
		select {
		case msgFromEventBus := <-eventBus:
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msgFromEventBus)); err != nil {
				fmt.Println("Error writing message:", err)
				return
			}
		case clientMsg, ok := <-clientMessages:
			if !ok {
				return // Exit the loop if clientMessages channel is closed
			}
			fmt.Println("Client message received:", clientMsg)
		}
	}

}

func main() {
	//read the redis connection string from the environment
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
		DB:   0,
	})

	router := gin.Default()
	router.GET("/", getMessage)
	router.GET("/details", func(c *gin.Context) {
		filename := c.Query("filename")
		result, err := processFile(filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})
	router.GET("/ws", websocketHandler)

	//start the redis subscription
	// messageChan := make(chan string)
	// go redisSubscribe("channel", messageChan)
	go redisSubscribe("channel")

	//start go http and websocket server
	router.Run("0.0.0.0:6000")
}
