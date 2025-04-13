package dgocacheler

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// BenchmarkAddMessage measures the performance of adding individual messages
func BenchmarkAddMessage(b *testing.B) {
	cache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(b.N)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cache.AddMessage("test-channel", messages[i])
	}
}

// BenchmarkAddMessageBatch measures the performance of adding messages in batches
func BenchmarkAddMessageBatch(b *testing.B) {
	cache := NewMessageCache(1000)
	batchSize := 100

	// We'll do b.N/batchSize iterations with batchSize messages each
	iterations := b.N / batchSize
	if iterations < 1 {
		iterations = 1
	}

	for i := 0; i < iterations; i++ {
		messages := TestHelpers.GenerateMessages(batchSize)

		b.ResetTimer()
		b.ReportAllocs()

		cache.AddMessages("test-channel", messages)
	}
}

// BenchmarkGetMessages measures the performance of retrieving messages
func BenchmarkGetMessages(b *testing.B) {
	cache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(1000)
	cache.AddMessages("test-channel", messages)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cache.GetMessages("test-channel")
	}
}

// BenchmarkGetMessagesLimit measures the performance of retrieving a limited number of messages
func BenchmarkGetMessagesLimit(b *testing.B) {
	cache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(1000)
	cache.AddMessages("test-channel", messages)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cache.GetMessagesLimit("test-channel", 100)
	}
}

// BenchmarkMultiChannelWrites measures the performance with concurrent writes to multiple channels
func BenchmarkMultiChannelWrites(b *testing.B) {
	cache := NewMessageCache(1000)
	numChannels := 10
	messagesPerChannel := b.N / numChannels

	if messagesPerChannel < 1 {
		messagesPerChannel = 1
	}

	// Pre-generate all messages
	channelMessages := make(map[string][]*discordgo.Message)
	for i := 0; i < numChannels; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		channelMessages[channelID] = TestHelpers.GenerateMessages(messagesPerChannel)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < numChannels; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		cache.AddMessages(channelID, channelMessages[channelID])
	}
}

// BenchmarkCircularBufferOverflow measures the performance of adding messages beyond capacity
func BenchmarkCircularBufferOverflow(b *testing.B) {
	// Small cache to ensure overflow
	cacheSize := 100
	cache := NewMessageCache(cacheSize)

	b.ResetTimer()
	b.ReportAllocs()

	// Add more messages than the cache size to trigger circular buffer behavior
	for i := 0; i < b.N; i++ {
		msg := &discordgo.Message{
			ID:      fmt.Sprintf("msg-%d", i),
			Content: fmt.Sprintf("Test message %d", i),
		}
		cache.AddMessage("test-channel", msg)
	}
}

// BenchmarkParallelReads measures the performance of concurrent reads
func BenchmarkParallelReads(b *testing.B) {
	cache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(1000)

	// Set up 10 channels with 100 messages each
	for i := 0; i < 10; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		cache.AddMessages(channelID, messages[:100])
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			channelID := fmt.Sprintf("channel-%d", counter%10)
			cache.GetMessages(channelID)
			counter++
		}
	})
}

// BenchmarkParallelReadWrite measures the performance of concurrent reads and writes
func BenchmarkParallelReadWrite(b *testing.B) {
	cache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(1000)

	// Set up initial data
	for i := 0; i < 10; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		cache.AddMessages(channelID, messages[:100])
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			// Alternate between reads and writes
			if counter%2 == 0 {
				// Read operation
				channelID := fmt.Sprintf("channel-%d", counter%10)
				cache.GetMessagesLimit(channelID, 50)
			} else {
				// Write operation
				channelID := fmt.Sprintf("channel-%d", counter%10)
				msg := &discordgo.Message{
					ID:      fmt.Sprintf("new-msg-%d", counter),
					Content: fmt.Sprintf("New message %d", counter),
				}
				cache.AddMessage(channelID, msg)
			}
			counter++
		}
	})
}

// BenchmarkGlobalCacheAccess measures the performance of accessing the global cache
func BenchmarkGlobalCacheAccess(b *testing.B) {
	// Reset the global cache before benchmarking
	globalCache = nil
	globalCacheOnce = sync.Once{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GetGlobalCache()
	}
}

// BenchmarkCompareOldVsNew compares old-style vs. new-style access
func BenchmarkCompareOldVsNew(b *testing.B) {
	// Test with channel-specific locks, reset globals first
	globalCache = nil
	globalCacheOnce = sync.Once{}

	// Ensure both are initialized first to avoid measuring init time
	_ = GetGlobalCache()
	_ = Cache

	messages := TestHelpers.GenerateMessages(1000)

	b.Run("OldStyleAccess", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N && i < len(messages); i++ {
			Cache.AddMessage("test-channel", messages[i])
		}
	})

	b.Run("NewStyleAccess", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N && i < len(messages); i++ {
			GetGlobalCache().AddMessage("test-channel", messages[i])
		}
	})
}

// BenchmarkSetMaxMessages measures the performance of changing the cache size
func BenchmarkSetMaxMessages(b *testing.B) {
	cache := NewMessageCache(1000)
	messages := TestHelpers.GenerateMessages(800)
	cache.AddMessages("test-channel", messages)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Alternate between increasing and decreasing
		if i%2 == 0 {
			cache.SetMaxMessages(1200)
		} else {
			cache.SetMaxMessages(500)
		}
	}
}
