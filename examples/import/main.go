package main

import (
	"fmt"

	"github.com/CreativeUnicorns/dgocacheler"
	"github.com/CreativeUnicorns/dgocacheler/examples/import/subpkg"
	"github.com/bwmarrin/discordgo"
)

func main() {
	// Set the max number of messages the already initialized global cache
	dgocacheler.Cache.SetMaxMessages(10)

	// Add some messages to the cache
	for i := 0; i < 10; i++ {
		msg := &discordgo.Message{ID: fmt.Sprintf("%d", i), Content: fmt.Sprintf("Ch 1. Message %d", i)}
		dgocacheler.Cache.AddMessage("channel1", msg)
	}

	// Add some messages to the cache for channel 2
	subpkg.AppendChannel2()

	// Retrieve messages from the cache for channel 1
	messages, _ := dgocacheler.Cache.GetMessages("channel1")
	for _, msg := range messages {
		fmt.Println(msg.Content)
	}

	// Retrieve messages from the cache for channel 2
	messages, _ = dgocacheler.Cache.GetMessages("channel2")
	for _, msg := range messages {
		fmt.Println(msg.Content)
	}
}
