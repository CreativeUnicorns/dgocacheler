package main

import (
	"fmt"

	"github.com/CreativeUnicorns/dgocacheler"
	"github.com/CreativeUnicorns/dgocacheler/examples/import/subpkg"
	"github.com/bwmarrin/discordgo"
)

func main() {
	// Get the global cache instance
	globalCache := dgocacheler.GetGlobalCache()

	// Set the max number of messages for the global cache
	err := globalCache.SetMaxMessages(10)
	if err != nil {
		fmt.Printf("Error setting max messages: %v\n", err)
		return
	}

	// Add some messages to the cache
	for i := 0; i < 10; i++ {
		msg := &discordgo.Message{ID: fmt.Sprintf("%d", i), Content: fmt.Sprintf("Ch 1. Message %d", i)}
		err := globalCache.AddMessage("channel1", msg)
		if err != nil {
			fmt.Printf("Error adding message: %v\n", err)
		}
	}

	// Add some messages to the cache for channel 2
	subpkg.AppendChannel2()

	// Retrieve messages from the cache for channel 1
	messages, err := globalCache.GetMessages("channel1")
	if err != nil {
		fmt.Printf("Error getting messages: %v\n", err)
	} else {
		fmt.Println("Channel 1 messages:")
		for _, msg := range messages {
			fmt.Println(msg.Content)
		}
	}

	// Retrieve messages from the cache for channel 2
	messages, err = globalCache.GetMessages("channel2")
	if err != nil {
		fmt.Printf("Error getting messages: %v\n", err)
	} else {
		fmt.Println("Channel 2 messages:")
		for _, msg := range messages {
			fmt.Println(msg.Content)
		}
	}
}
