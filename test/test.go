package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, World!")
	
	// This line has an error - os.Stdout doesn't exist
	os.Stdout.WriteString("test")
	
	// Test variables for completion
	var myVariable int = 42
	var myString string = "test"
	
	// Try to use variables
	fmt.Printf("Number: %d, String: %s\n", myVariable, myString)
}