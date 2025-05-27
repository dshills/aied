package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	// Variables for testing completion
	message := "Hello, World!"
	numbers := []int{1, 2, 3, 4, 5}
	now := time.Now()
	
	// Test string methods (should trigger completion after '.')
	upperMessage := strings.ToUpper(message)
	fmt.Println(upperMessage)
	
	// Test slice methods
	length := len(numbers)
	fmt.Printf("Length: %d\n", length)
	
	// Test time methods
	formatted := now.Format("2006-01-02 15:04:05")
	fmt.Println(formatted)
	
	// Test with incomplete typing for completion
	fmt.
	strings.
	time.
}

// Function for testing go-to-definition
func helper(x int) string {
	return fmt.Sprintf("Number: %d", x)
}

// Function that calls helper (for testing references)
func caller() {
	result := helper(42)
	fmt.Println(result)
}