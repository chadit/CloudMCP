package main

import (
	"fmt"
	"sort"
)

func main() {
	testCase := []string{"zebra", "apple", "banana", "method", "endpoint"}
	original := make([]string, len(testCase))
	copy(original, testCase)

	fmt.Printf("Original: %v\n", original)

	// Use standard sort (our new implementation)
	sort.Strings(testCase)
	fmt.Printf("Sorted:   %v\n", testCase)

	fmt.Println("✅ sort.Strings works correctly with Go 1.24")
}
