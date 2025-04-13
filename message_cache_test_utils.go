package dgocacheler

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// TestHelpers contains utility functions for testing and benchmarking
var TestHelpers = struct {
	// GenerateMessages creates n test messages
	GenerateMessages func(n int) []*discordgo.Message
}{
	GenerateMessages: func(n int) []*discordgo.Message {
		messages := make([]*discordgo.Message, n)
		for i := 0; i < n; i++ {
			messages[i] = &discordgo.Message{
				ID:      fmt.Sprintf("msg-%d", i),
				Content: fmt.Sprintf("This is test message %d with some content to simulate a real message.", i),
				Author: &discordgo.User{
					ID:       fmt.Sprintf("user-%d", i%100),
					Username: fmt.Sprintf("User%d", i%100),
				},
				ChannelID: "test-channel",
			}
		}
		return messages
	},
}
