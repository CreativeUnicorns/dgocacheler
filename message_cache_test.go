package dgocacheler

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNewMessageCache(t *testing.T) {
	cache := NewMessageCache(10)
	if cache == nil {
		t.Error("NewMessageCache did not create a cache instance.")
	}
	if cache != nil && len(cache.messages) != 0 {
		t.Error("New cache should be empty.")
	}
}

func TestAddMessage(t *testing.T) {
	cache := NewMessageCache(5)
	msg := &discordgo.Message{ID: "1", Content: "Hello, World!"}
	cache.AddMessage("channel1", msg)

	if msgs, ok := cache.GetMessages("channel1"); !ok || len(msgs) != 1 {
		t.Error("AddMessage failed to add message to the cache.")
	}
}

func TestSetMaxMessages(t *testing.T) {
	cache := NewMessageCache(5)
	for i := 0; i < 5; i++ {
		cache.AddMessage("channel1", &discordgo.Message{ID: fmt.Sprint(i)})
	}

	// Increase the cache size
	cache.SetMaxMessages(10)
	if cache.maxMessages != 10 {
		t.Errorf("SetMaxMessages did not correctly set the new maximum size, got %d", cache.maxMessages)
	}

	// Add more messages to fill increased size
	for i := 5; i < 10; i++ {
		cache.AddMessage("channel1", &discordgo.Message{ID: fmt.Sprint(i)})
	}
	if msgs, ok := cache.GetMessages("channel1"); !ok || len(msgs) != 10 {
		t.Error("SetMaxMessages failed to handle increased size correctly.")
	}

	// Reduce the cache size
	cache.SetMaxMessages(3)
	if msgs, ok := cache.GetMessages("channel1"); !ok || len(msgs) != 3 {
		t.Error("SetMaxMessages failed to reduce cache size and purge old messages.")
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := NewMessageCache(100)
	// Simulate concurrent access
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(id int) {
			cache.AddMessage("channel1", &discordgo.Message{ID: fmt.Sprint(id)})
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 100; i++ {
		<-done
	}

	if msgs, ok := cache.GetMessages("channel1"); !ok || len(msgs) != 100 {
		t.Errorf("Expected 100 messages, got %d", len(msgs))
	}
}
