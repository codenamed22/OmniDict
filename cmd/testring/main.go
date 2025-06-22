package main

import (
	"fmt"
	"omnidict/internal/ring"
)

func main() {
	// Create a new hash ring with 3 virtual nodes per physical node
	hashRing := ring.NewHashRing(3)

	// Add some nodes
	hashRing.AddNode("nodeA")
	hashRing.AddNode("nodeB")
	hashRing.AddNode("nodeC")

	// Add some key-value pairs
	keys := []string{"apple", "banana", "cherry", "date", "elderberry"}

	for _, key := range keys {
		value := key + "_value"
		hashRing.Set(key, value)
	}

	// Retrieve the values
	for _, key := range keys {
		val, ok := hashRing.Get(key)
		if ok {
			fmt.Printf("Get('%s') = '%s'\n", key, val)
		} else {
			fmt.Printf("Key '%s' not found\n", key)
		}
	}

	// Print the ring status to show how keys are distributed
	hashRing.PrintRingStatus()
}
