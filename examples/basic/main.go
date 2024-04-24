package main

import (
	"fmt"

	"github.com/CreativeUnicorns/dgocacheler"
	"github.com/bwmarrin/discordgo"
)

func main() {
	// Create a new message cache with a max size of 10 messages
	cache := dgocacheler.NewMessageCache(10)

	// Add some messages to the cache
	for i := 0; i < 10; i++ {
		msg := &discordgo.Message{ID: fmt.Sprintf("%d", i), Content: fmt.Sprintf("Message %d", i)}
		cache.AddMessage("channel1", msg)
	}

	// Retrieve messages from the cache
	messages, _ := cache.GetMessages("channel1")
	for _, msg := range messages {
		fmt.Println(msg.Content)
	}
}
