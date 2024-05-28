package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type details struct {
	Lines  int `json:"lines"`
	Words  int `json:"words"`
	Letter int `json:"letters"`
}

type message struct {
	Message string `json:"message"`
}

func getDetails(c *gin.Context) {
	filename := "go.mod"
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines, words, letters := 0, 0, 0
	for scanner.Scan() {
		lines++
		words += len(strings.Fields(scanner.Text()))
		letters += len(scanner.Text())
	}

	fmt.Println("Lines:", lines)
	fmt.Println("Words:", words)
	fmt.Println("Letters:", letters)
	result := details{Lines: lines, Words: words, Letter: letters}

	c.JSON(http.StatusOK, result)
}

func getMessage(c *gin.Context) {
	c.JSON(http.StatusOK, message{Message: "PONG"})
}

func main() {
	router := gin.Default()
	router.GET("/", getMessage)
	router.GET("/details", getDetails)

	router.Run("localhost:6000")
}
