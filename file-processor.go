package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	filename := "sample.txt"
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
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
}
