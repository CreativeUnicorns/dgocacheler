// Package cacheler provides a concurrency-safe cache for storing Discord messages by channel.

package dgocacheler

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

// MessageCache holds Discord messages organized by channel ID. It supports concurrent access.
type MessageCache struct {
	sync.RWMutex                                 // Embedding RWMutex to provide locking
	messages     map[string][]*discordgo.Message // messages maps channel IDs to slice of messages
	maxMessages  int                             // maxMessages defines the max number of messages per channel
}

// NewMessageCache creates a new MessageCache with a specified maximum number of messages per channel.
func NewMessageCache(maxMessages int) *MessageCache {
	return &MessageCache{
		messages:    make(map[string][]*discordgo.Message),
		maxMessages: maxMessages,
	}
}

// AddMessage adds a single message to the cache for a specific channel.
func (c *MessageCache) AddMessage(channelID string, message *discordgo.Message) {
	c.Lock()
	defer c.Unlock()
	c.addMessageInternal(channelID, message)
}

// AddMessages adds multiple messages to the cache for a specific channel.
func (c *MessageCache) AddMessages(channelID string, messages []*discordgo.Message) {
	c.Lock()
	defer c.Unlock()
	for _, message := range messages {
		c.addMessageInternal(channelID, message)
	}
}

// addMessageInternal is an unexported helper function that handles the actual addition of messages to the cache.
func (c *MessageCache) addMessageInternal(channelID string, message *discordgo.Message) {
	if _, ok := c.messages[channelID]; !ok {
		c.messages[channelID] = []*discordgo.Message{}
	}
	c.messages[channelID] = append(c.messages[channelID], message)
	if len(c.messages[channelID]) > c.maxMessages {
		c.messages[channelID] = c.messages[channelID][1:]
	}
}

// GetMessages retrieves all messages for a given channel from the cache
func (c *MessageCache) GetMessages(channelID string) ([]*discordgo.Message, bool) {
	c.RLock()
	defer c.RUnlock()
	msgs, ok := c.messages[channelID]
	return msgs, ok
}

// GetMessagesLimit retrieves up to a specified number of recent messages for a given channel.
func (c *MessageCache) GetMessagesLimit(channelID string, limit int) ([]*discordgo.Message, bool) {
	c.RLock()
	defer c.RUnlock()
	msgs, ok := c.messages[channelID]
	if !ok || len(msgs) == 0 {
		return nil, false
	}
	start := len(msgs) - limit
	if start < 0 {
		start = 0
	}
	return msgs[start:], true
}

// SetMaxMessages sets the maximum number of messages to store per channel in the cache.
func (c *MessageCache) SetMaxMessages(maxMessages int) {
	c.Lock()
	defer c.Unlock()
	c.maxMessages = maxMessages
	for channelID, messages := range c.messages {
		if len(messages) > maxMessages {
			c.messages[channelID] = messages[len(messages)-maxMessages:]
		}
	}
}

// Global cache
var Cache = NewMessageCache(100)
