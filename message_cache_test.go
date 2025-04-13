package dgocacheler

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNewMessageCache(t *testing.T) {
	cache := NewMessageCache(10)
	if cache == nil {
		t.Error("NewMessageCache did not create a cache instance.")
	}
	if cache != nil && len(cache.channels) != 0 {
		t.Error("New cache should be empty.")
	}
}

func TestNewMessageCacheWithInvalidSize(t *testing.T) {
	// Test with invalid size, should default to 100
	cache := NewMessageCache(-5)
	if cache.maxMessages != 100 {
		t.Errorf("Expected default max messages of 100, got %d", cache.maxMessages)
	}
}

func TestAddMessage(t *testing.T) {
	cache := NewMessageCache(5)
	msg := &discordgo.Message{ID: "1", Content: "Hello, World!"}

	err := cache.AddMessage("channel1", msg)
	if err != nil {
		t.Errorf("AddMessage returned unexpected error: %v", err)
	}

	msgs, err := cache.GetMessages("channel1")
	if err != nil {
		t.Errorf("GetMessages returned unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Errorf("Expected 1 message, got %d", len(msgs))
	}
}

func TestAddMessageWithNilMessage(t *testing.T) {
	cache := NewMessageCache(5)
	err := cache.AddMessage("channel1", nil)
	if err != ErrNilMessage {
		t.Errorf("Expected ErrNilMessage, got %v", err)
	}
}

func TestAddMessageWithEmptyChannelID(t *testing.T) {
	cache := NewMessageCache(5)
	msg := &discordgo.Message{ID: "1", Content: "Hello, World!"}
	err := cache.AddMessage("", msg)
	if err != ErrInvalidChannel {
		t.Errorf("Expected ErrInvalidChannel, got %v", err)
	}
}

func TestAddDuplicateMessage(t *testing.T) {
	cache := NewMessageCache(5)
	msg := &discordgo.Message{ID: "1", Content: "Hello, World!"}

	// Add the message first time
	err := cache.AddMessage("channel1", msg)
	if err != nil {
		t.Errorf("First AddMessage returned unexpected error: %v", err)
	}

	// Add the same message again
	err = cache.AddMessage("channel1", msg)
	if err != nil {
		t.Errorf("Second AddMessage returned unexpected error: %v", err)
	}

	// Verify only one message was added
	msgs, err := cache.GetMessages("channel1")
	if err != nil {
		t.Errorf("GetMessages returned unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Errorf("Expected 1 message after duplicate add, got %d", len(msgs))
	}
}

func TestAddMessages(t *testing.T) {
	cache := NewMessageCache(5)
	messages := []*discordgo.Message{
		{ID: "1", Content: "First message"},
		{ID: "2", Content: "Second message"},
	}

	err := cache.AddMessages("channel1", messages)
	if err != nil {
		t.Errorf("AddMessages returned unexpected error: %v", err)
	}

	msgs, err := cache.GetMessages("channel1")
	if err != nil {
		t.Errorf("GetMessages returned unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(msgs))
	}
}

func TestAddMessagesWithNilMessage(t *testing.T) {
	cache := NewMessageCache(5)
	messages := []*discordgo.Message{
		{ID: "1", Content: "First message"},
		nil,
		{ID: "2", Content: "Second message"},
	}

	err := cache.AddMessages("channel1", messages)
	if err != nil {
		t.Errorf("AddMessages should skip nil messages without error, got: %v", err)
	}

	msgs, err := cache.GetMessages("channel1")
	if err != nil {
		t.Errorf("GetMessages returned unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("Expected 2 messages (skipping nil), got %d", len(msgs))
	}
}

func TestGetMessagesForNonExistentChannel(t *testing.T) {
	cache := NewMessageCache(5)
	_, err := cache.GetMessages("nonexistent")
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss, got %v", err)
	}
}

func TestGetMessagesLimit(t *testing.T) {
	cache := NewMessageCache(10)

	// Add 5 messages
	for i := 0; i < 5; i++ {
		cache.AddMessage("channel1", &discordgo.Message{ID: fmt.Sprintf("%d", i)})
	}

	// Get 3 most recent messages
	msgs, err := cache.GetMessagesLimit("channel1", 3)
	if err != nil {
		t.Errorf("GetMessagesLimit returned unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(msgs))
	}

	// Verify most recent messages (ID 2, 3, 4)
	for i := 0; i < 3; i++ {
		expectedID := fmt.Sprintf("%d", i+2)
		if msgs[i].ID != expectedID {
			t.Errorf("Expected message ID %s at position %d, got %s", expectedID, i, msgs[i].ID)
		}
	}
}

func TestGetMessagesLimitWithInvalidLimit(t *testing.T) {
	cache := NewMessageCache(5)
	_, err := cache.GetMessagesLimit("channel1", 0)
	if err != ErrInvalidLimit {
		t.Errorf("Expected ErrInvalidLimit, got %v", err)
	}
}

func TestSetMaxMessages(t *testing.T) {
	cache := NewMessageCache(5)
	for i := 0; i < 5; i++ {
		cache.AddMessage("channel1", &discordgo.Message{ID: fmt.Sprintf("%d", i)})
	}

	// Increase the cache size
	err := cache.SetMaxMessages(10)
	if err != nil {
		t.Errorf("SetMaxMessages returned unexpected error: %v", err)
	}
	if cache.maxMessages != 10 {
		t.Errorf("SetMaxMessages did not correctly set the new maximum size, got %d", cache.maxMessages)
	}

	// Add more messages to fill increased size
	for i := 5; i < 10; i++ {
		cache.AddMessage("channel1", &discordgo.Message{ID: fmt.Sprintf("%d", i)})
	}
	msgs, err := cache.GetMessages("channel1")
	if err != nil {
		t.Errorf("GetMessages returned unexpected error: %v", err)
	}
	if len(msgs) != 10 {
		t.Errorf("SetMaxMessages failed to handle increased size correctly. Expected 10, got %d", len(msgs))
	}

	// Reduce the cache size
	err = cache.SetMaxMessages(3)
	if err != nil {
		t.Errorf("SetMaxMessages returned unexpected error: %v", err)
	}
	msgs, err = cache.GetMessages("channel1")
	if err != nil {
		t.Errorf("GetMessages returned unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Errorf("SetMaxMessages failed to reduce cache size. Expected 3, got %d", len(msgs))
	}

	// Check that the most recent messages were kept (IDs 7, 8, 9)
	expectedIDs := []string{"7", "8", "9"}
	for i, expectedID := range expectedIDs {
		if msgs[i].ID != expectedID {
			t.Errorf("Expected message ID %s at position %d, got %s", expectedID, i, msgs[i].ID)
		}
	}
}

func TestSetMaxMessagesWithInvalidSize(t *testing.T) {
	cache := NewMessageCache(5)
	err := cache.SetMaxMessages(-10)
	if err != ErrInvalidLimit {
		t.Errorf("Expected ErrInvalidLimit, got %v", err)
	}
}

func TestClearChannel(t *testing.T) {
	cache := NewMessageCache(5)
	for i := 0; i < 3; i++ {
		cache.AddMessage("channel1", &discordgo.Message{ID: fmt.Sprintf("%d", i)})
	}

	err := cache.ClearChannel("channel1")
	if err != nil {
		t.Errorf("ClearChannel returned unexpected error: %v", err)
	}

	msgs, err := cache.GetMessages("channel1")
	if err != nil {
		t.Errorf("GetMessages returned unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("ClearChannel failed to clear messages. Expected 0, got %d", len(msgs))
	}
}

func TestClearNonExistentChannel(t *testing.T) {
	cache := NewMessageCache(5)
	err := cache.ClearChannel("nonexistent")
	if err != nil {
		t.Errorf("ClearChannel should not return error for non-existent channel, got: %v", err)
	}
}

func TestCircularBufferBehavior(t *testing.T) {
	cache := NewMessageCache(3)

	// Add more messages than the capacity
	for i := 0; i < 5; i++ {
		cache.AddMessage("channel1", &discordgo.Message{ID: fmt.Sprintf("%d", i)})
	}

	msgs, err := cache.GetMessages("channel1")
	if err != nil {
		t.Errorf("GetMessages returned unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Errorf("Expected 3 messages (limited by capacity), got %d", len(msgs))
	}

	// Verify only most recent messages are kept (IDs 2, 3, 4)
	expectedIDs := []string{"2", "3", "4"}
	for i, expectedID := range expectedIDs {
		if msgs[i].ID != expectedID {
			t.Errorf("Expected message ID %s at position %d, got %s", expectedID, i, msgs[i].ID)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := NewMessageCache(100)

	// Simulate concurrent access from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := &discordgo.Message{ID: fmt.Sprintf("%d", id), Content: fmt.Sprintf("Message %d", id)}
			err := cache.AddMessage("channel1", msg)
			if err != nil {
				t.Errorf("AddMessage in goroutine returned error: %v", err)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify all messages were added
	msgs, err := cache.GetMessages("channel1")
	if err != nil {
		t.Errorf("GetMessages returned unexpected error: %v", err)
	}
	if len(msgs) != 100 {
		t.Errorf("Expected 100 messages after concurrent adds, got %d", len(msgs))
	}
}

func TestGlobalCache(t *testing.T) {
	// Clear any existing global cache for testing
	globalCache = nil
	globalCacheOnce = sync.Once{}

	// Get the global cache
	cache1 := GetGlobalCache()
	if cache1 == nil {
		t.Error("GetGlobalCache returned nil")
	}

	// Add a message to the global cache
	err := cache1.AddMessage("global-channel", &discordgo.Message{ID: "global-1"})
	if err != nil {
		t.Errorf("AddMessage to global cache returned error: %v", err)
	}

	// Get the global cache again (should be the same instance)
	cache2 := GetGlobalCache()
	if cache2 != cache1 {
		t.Error("GetGlobalCache returned a different instance")
	}

	// Verify the message is in the global cache
	msgs, err := cache2.GetMessages("global-channel")
	if err != nil {
		t.Errorf("GetMessages from global cache returned error: %v", err)
	}
	if len(msgs) != 1 || msgs[0].ID != "global-1" {
		t.Error("Message not correctly stored in global cache")
	}
}
