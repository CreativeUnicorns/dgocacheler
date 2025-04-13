package dgocacheler

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// These types simulate the old v0.1.1 implementation
type OldMessageCache struct {
	sync.RWMutex
	messages    map[string][]*discordgo.Message
	maxMessages int
}

func NewOldMessageCache(maxMessages int) *OldMessageCache {
	return &OldMessageCache{
		messages:    make(map[string][]*discordgo.Message),
		maxMessages: maxMessages,
	}
}

func (c *OldMessageCache) AddMessage(channelID string, message *discordgo.Message) {
	c.Lock()
	defer c.Unlock()
	c.addMessageInternal(channelID, message)
}

func (c *OldMessageCache) AddMessages(channelID string, messages []*discordgo.Message) {
	c.Lock()
	defer c.Unlock()
	for _, message := range messages {
		c.addMessageInternal(channelID, message)
	}
}

func (c *OldMessageCache) addMessageInternal(channelID string, message *discordgo.Message) {
	if _, ok := c.messages[channelID]; !ok {
		c.messages[channelID] = []*discordgo.Message{}
	}
	c.messages[channelID] = append(c.messages[channelID], message)
	if len(c.messages[channelID]) > c.maxMessages {
		c.messages[channelID] = c.messages[channelID][1:]
	}
}

func (c *OldMessageCache) GetMessages(channelID string) ([]*discordgo.Message, bool) {
	c.RLock()
	defer c.RUnlock()
	msgs, ok := c.messages[channelID]
	return msgs, ok
}

func (c *OldMessageCache) GetMessagesLimit(channelID string, limit int) ([]*discordgo.Message, bool) {
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

func (c *OldMessageCache) SetMaxMessages(maxMessages int) {
	c.Lock()
	defer c.Unlock()
	c.maxMessages = maxMessages
	for channelID, messages := range c.messages {
		if len(messages) > maxMessages {
			c.messages[channelID] = messages[len(messages)-maxMessages:]
		}
	}
}

// BenchmarkComparison_AddMessage compares adding messages between old and new implementations
func BenchmarkComparison_AddMessage(b *testing.B) {
	oldCache := NewOldMessageCache(1000)
	newCache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(b.N)

	b.Run("OldImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N && i < len(messages); i++ {
			oldCache.AddMessage("test-channel", messages[i])
		}
	})

	b.Run("NewImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N && i < len(messages); i++ {
			newCache.AddMessage("test-channel", messages[i])
		}
	})
}

// BenchmarkComparison_AddMessageBatch compares adding message batches
func BenchmarkComparison_AddMessageBatch(b *testing.B) {
	oldCache := NewOldMessageCache(1000)
	newCache := NewMessageCache(1000)
	batchSize := 100

	// We'll do b.N/batchSize iterations with batchSize messages each
	iterations := b.N / batchSize
	if iterations < 1 {
		iterations = 1
	}

	messages := TestHelpers.GenerateMessages(batchSize * iterations)

	b.Run("OldImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < iterations; i++ {
			start := i * batchSize
			end := start + batchSize
			if end > len(messages) {
				end = len(messages)
			}
			oldCache.AddMessages("test-channel", messages[start:end])
		}
	})

	b.Run("NewImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < iterations; i++ {
			start := i * batchSize
			end := start + batchSize
			if end > len(messages) {
				end = len(messages)
			}
			newCache.AddMessages("test-channel", messages[start:end])
		}
	})
}

// BenchmarkComparison_GetMessages compares retrieving messages
func BenchmarkComparison_GetMessages(b *testing.B) {
	oldCache := NewOldMessageCache(1000)
	newCache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(1000)

	oldCache.AddMessages("test-channel", messages)
	newCache.AddMessages("test-channel", messages)

	b.Run("OldImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = oldCache.GetMessages("test-channel")
		}
	})

	b.Run("NewImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = newCache.GetMessages("test-channel")
		}
	})
}

// BenchmarkComparison_GetMessagesLimit compares retrieving limited messages
func BenchmarkComparison_GetMessagesLimit(b *testing.B) {
	oldCache := NewOldMessageCache(1000)
	newCache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(1000)

	oldCache.AddMessages("test-channel", messages)
	newCache.AddMessages("test-channel", messages)

	b.Run("OldImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = oldCache.GetMessagesLimit("test-channel", 100)
		}
	})

	b.Run("NewImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = newCache.GetMessagesLimit("test-channel", 100)
		}
	})
}

// BenchmarkComparison_MultiChannel compares performance with multiple channels
func BenchmarkComparison_MultiChannel(b *testing.B) {
	oldCache := NewOldMessageCache(1000)
	newCache := NewMessageCache(1000)
	numChannels := 10
	messagesPerChannel := 100

	// Pre-generate all messages
	allMessages := TestHelpers.GenerateMessages(numChannels * messagesPerChannel)

	b.Run("OldImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			for j := 0; j < numChannels; j++ {
				channelID := fmt.Sprintf("channel-%d", j)
				start := j * messagesPerChannel
				end := start + messagesPerChannel
				if end > len(allMessages) {
					end = len(allMessages)
				}
				oldCache.AddMessages(channelID, allMessages[start:end])
			}
		}
	})

	b.Run("NewImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			for j := 0; j < numChannels; j++ {
				channelID := fmt.Sprintf("channel-%d", j)
				start := j * messagesPerChannel
				end := start + messagesPerChannel
				if end > len(allMessages) {
					end = len(allMessages)
				}
				newCache.AddMessages(channelID, allMessages[start:end])
			}
		}
	})
}

// BenchmarkComparison_CircularBuffer compares the circular buffer behavior
func BenchmarkComparison_CircularBuffer(b *testing.B) {
	oldCache := NewOldMessageCache(100)
	newCache := NewMessageCache(100)

	b.Run("OldImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			msg := &discordgo.Message{
				ID:      fmt.Sprintf("msg-%d", i),
				Content: fmt.Sprintf("Test message %d", i),
			}
			oldCache.AddMessage("test-channel", msg)
		}
	})

	b.Run("NewImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			msg := &discordgo.Message{
				ID:      fmt.Sprintf("msg-%d", i),
				Content: fmt.Sprintf("Test message %d", i),
			}
			newCache.AddMessage("test-channel", msg)
		}
	})
}

// BenchmarkComparison_ParallelReads compares parallel read performance
func BenchmarkComparison_ParallelReads(b *testing.B) {
	oldCache := NewOldMessageCache(1000)
	newCache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(1000)

	// Set up 10 channels with messages in both caches
	for i := 0; i < 10; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		oldCache.AddMessages(channelID, messages[:100])
		newCache.AddMessages(channelID, messages[:100])
	}

	b.Run("OldImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				channelID := fmt.Sprintf("channel-%d", counter%10)
				oldCache.GetMessages(channelID)
				counter++
			}
		})
	})

	b.Run("NewImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				channelID := fmt.Sprintf("channel-%d", counter%10)
				newCache.GetMessages(channelID)
				counter++
			}
		})
	})
}

// BenchmarkComparison_ParallelReadWrite compares parallel read/write performance
func BenchmarkComparison_ParallelReadWrite(b *testing.B) {
	oldCache := NewOldMessageCache(1000)
	newCache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(1000)

	// Set up initial data in both caches
	for i := 0; i < 10; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		oldCache.AddMessages(channelID, messages[:100])
		newCache.AddMessages(channelID, messages[:100])
	}

	b.Run("OldImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				// Alternate between reads and writes
				if counter%2 == 0 {
					// Read operation
					channelID := fmt.Sprintf("channel-%d", counter%10)
					oldCache.GetMessagesLimit(channelID, 50)
				} else {
					// Write operation
					channelID := fmt.Sprintf("channel-%d", counter%10)
					msg := &discordgo.Message{
						ID:      fmt.Sprintf("new-msg-%d", counter),
						Content: fmt.Sprintf("New message %d", counter),
					}
					oldCache.AddMessage(channelID, msg)
				}
				counter++
			}
		})
	})

	b.Run("NewImplementation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				// Alternate between reads and writes
				if counter%2 == 0 {
					// Read operation
					channelID := fmt.Sprintf("channel-%d", counter%10)
					newCache.GetMessagesLimit(channelID, 50)
				} else {
					// Write operation
					channelID := fmt.Sprintf("channel-%d", counter%10)
					msg := &discordgo.Message{
						ID:      fmt.Sprintf("new-msg-%d", counter),
						Content: fmt.Sprintf("New message %d", counter),
					}
					newCache.AddMessage(channelID, msg)
				}
				counter++
			}
		})
	})
}
