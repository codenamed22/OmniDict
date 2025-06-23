package main

import (
	"fmt"
	"omnidict/internal/store"
)

func main() {
	// Create a new store
	kv := store.NewStore()

	// Put some key-value pairs
	proto.Put("name", "Eron")
	proto.Put("language", "GoLang")

	// Get values
	if val, ok := proto.Get("name"); ok {
		fmt.Println("Key: name →", val)
	} else {
		fmt.Println("Key 'name' not found")
	}

	if val, ok := proto.Get("language"); ok {
		fmt.Println("Key: language →", val)
	} else {
		fmt.Println("Key 'language' not found")
	}

	// Try a missing key
	if val, ok := proto.Get("age"); ok {
		fmt.Println("Key: age →", val)
	} else {
		fmt.Println("Key 'age' not found")
	}
}
