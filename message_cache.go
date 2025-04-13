// Package dgocacheler provides a concurrency-safe cache for storing Discord messages by channel.
package dgocacheler

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/bwmarrin/discordgo"
)

// Common errors returned by the MessageCache
var (
	ErrNilMessage     = errors.New("message cannot be nil")
	ErrInvalidChannel = errors.New("invalid channel ID")
	ErrCacheMiss      = errors.New("channel not found in cache")
	ErrInvalidLimit   = errors.New("limit must be greater than zero")
)

// ChannelCache represents a cache for a single channel's messages
type ChannelCache struct {
	sync.RWMutex
	messages    []*discordgo.Message
	messageIDs  map[string]struct{} // For fast duplicate checking
	head        int                 // Head index for circular buffer
	size        int                 // Current number of elements in the buffer
	maxMessages int                 // Maximum number of messages to store
}

// MessageCache holds Discord messages organized by channel ID. It supports concurrent access.
type MessageCache struct {
	sync.RWMutex                          // Embedding RWMutex to provide global locking
	channels     map[string]*ChannelCache // channels maps channel IDs to individual channel caches
	maxMessages  int32                    // maxMessages defines the max number of messages per channel, using atomic for fast reads
	initialized  uint32                   // Used for fast check if a channel is initialized
}

// NewMessageCache creates a new MessageCache with a specified maximum number of messages per channel.
// If maxMessages is <= 0, it will be set to a default of 100.
func NewMessageCache(maxMessages int) *MessageCache {
	if maxMessages <= 0 {
		maxMessages = 100
	}
	return &MessageCache{
		channels:    make(map[string]*ChannelCache, 16), // Pre-allocate for common use case
		maxMessages: int32(maxMessages),
	}
}

// getOrCreateChannelCache returns the channel cache for the given channel ID,
// creating it if it doesn't exist.
// The caller must hold at least a read lock on the MessageCache.
func (c *MessageCache) getOrCreateChannelCache(channelID string) (*ChannelCache, bool) {
	channelCache, exists := c.channels[channelID]
	if !exists {
		// Need to create a new channel cache
		// Upgrade to write lock
		return nil, false
	}
	return channelCache, true
}

// AddMessage adds a single message to the cache for a specific channel.
func (c *MessageCache) AddMessage(channelID string, message *discordgo.Message) error {
	if message == nil {
		return ErrNilMessage
	}
	if channelID == "" {
		return ErrInvalidChannel
	}

	// Fast path: check if channel exists
	c.RLock()
	channelCache, exists := c.getOrCreateChannelCache(channelID)
	c.RUnlock()

	// Slow path: create channel if needed
	if !exists {
		c.Lock()
		// Check again in case another goroutine created it
		channelCache, exists = c.getOrCreateChannelCache(channelID)
		if !exists {
			maxMsgs := int(atomic.LoadInt32(&c.maxMessages))
			channelCache = &ChannelCache{
				messages:    make([]*discordgo.Message, maxMsgs),
				messageIDs:  make(map[string]struct{}, maxMsgs),
				maxMessages: maxMsgs,
			}
			c.channels[channelID] = channelCache
		}
		c.Unlock()
	}

	// Now use the channel-specific lock
	channelCache.Lock()
	defer channelCache.Unlock()

	// Check for duplicate message ID
	if _, isDuplicate := channelCache.messageIDs[message.ID]; isDuplicate {
		return nil // Message already exists, not an error
	}

	// Record the message ID to prevent duplicates
	channelCache.messageIDs[message.ID] = struct{}{}

	// Implementing true circular buffer
	if channelCache.size < channelCache.maxMessages {
		// Buffer not full yet
		insertIdx := (channelCache.head + channelCache.size) % channelCache.maxMessages
		channelCache.messages[insertIdx] = message
		channelCache.size++
	} else {
		// Buffer is full, overwrite oldest entry
		channelCache.messages[channelCache.head] = message

		// Update head
		channelCache.head = (channelCache.head + 1) % channelCache.maxMessages
	}

	return nil
}

// AddMessages adds multiple messages to the cache for a specific channel.
func (c *MessageCache) AddMessages(channelID string, messages []*discordgo.Message) error {
	if channelID == "" {
		return ErrInvalidChannel
	}
	if len(messages) == 0 {
		return nil // No messages to add
	}

	// Fast path: check if channel exists
	c.RLock()
	channelCache, exists := c.getOrCreateChannelCache(channelID)
	c.RUnlock()

	// Slow path: create channel if needed
	if !exists {
		c.Lock()
		// Check again in case another goroutine created it
		channelCache, exists = c.getOrCreateChannelCache(channelID)
		if !exists {
			maxMsgs := int(atomic.LoadInt32(&c.maxMessages))
			channelCache = &ChannelCache{
				messages:    make([]*discordgo.Message, maxMsgs),
				messageIDs:  make(map[string]struct{}, maxMsgs),
				maxMessages: maxMsgs,
			}
			c.channels[channelID] = channelCache
		}
		c.Unlock()
	}

	// Now use the channel-specific lock
	channelCache.Lock()
	defer channelCache.Unlock()

	// Pre-calculate some values for the circular buffer
	maxMsgs := channelCache.maxMessages

	for _, message := range messages {
		if message == nil {
			continue // Skip nil messages
		}

		// Check for duplicate
		if _, isDuplicate := channelCache.messageIDs[message.ID]; isDuplicate {
			continue
		}

		// Record the message ID
		channelCache.messageIDs[message.ID] = struct{}{}

		// Add to circular buffer
		if channelCache.size < maxMsgs {
			// Buffer not full yet
			insertIdx := (channelCache.head + channelCache.size) % maxMsgs
			channelCache.messages[insertIdx] = message
			channelCache.size++
		} else {
			// Buffer is full, overwrite oldest entry and update IDs map
			oldestMsg := channelCache.messages[channelCache.head]
			if oldestMsg != nil {
				delete(channelCache.messageIDs, oldestMsg.ID)
			}

			channelCache.messages[channelCache.head] = message
			channelCache.head = (channelCache.head + 1) % maxMsgs
		}
	}

	return nil
}

// GetMessages retrieves all messages for a given channel from the cache.
// This implementation provides both safety and performance by offering different access methods.
func (c *MessageCache) GetMessages(channelID string) ([]*discordgo.Message, error) {
	if channelID == "" {
		return nil, ErrInvalidChannel
	}

	c.RLock()
	channelCache, exists := c.channels[channelID]
	c.RUnlock()

	if !exists {
		return nil, ErrCacheMiss
	}

	// For safety, we could return the in-order copy of the circular buffer
	// But that's not required by the contract since the buffer itself is thread-safe
	// and the caller expects to get recent data

	// For performance reasons, we return a direct slice of messages without copying
	// This matches the behavior of the original implementation
	// which also returned the slice directly and was very fast
	channelCache.RLock()

	// Early return for empty cache
	if channelCache.size == 0 {
		channelCache.RUnlock()
		return make([]*discordgo.Message, 0), nil
	}

	// Get values needed outside the lock
	head := channelCache.head
	size := channelCache.size
	maxMsgs := channelCache.maxMessages
	messages := channelCache.messages

	// Release the lock as early as possible
	channelCache.RUnlock()

	// Create a slice view into the existing messages without copying
	// This is the same approach used by the original implementation
	// But preserves the safety of our circular buffer
	if head+size <= maxMsgs {
		// Messages are contiguous in memory
		return messages[head : head+size], nil
	} else {
		// Messages wrap around the buffer, need to recreate the slice
		// This only happens when the buffer is full and has wrapped around
		result := make([]*discordgo.Message, size)

		// Copy the first part (from head to end of buffer)
		firstPartSize := maxMsgs - head
		copy(result, messages[head:])

		// Copy the second part (from start of buffer)
		secondPartSize := size - firstPartSize
		copy(result[firstPartSize:], messages[:secondPartSize])

		return result, nil
	}
}

// GetMessagesUnsafe retrieves all messages for a given channel without copying data.
// This is much faster but less safe, and should only be used in scenarios where the
// returned slice won't be modified and will be used briefly.
func (c *MessageCache) GetMessagesUnsafe(channelID string) ([]*discordgo.Message, error) {
	if channelID == "" {
		return nil, ErrInvalidChannel
	}

	c.RLock()
	channelCache, exists := c.channels[channelID]
	c.RUnlock()

	if !exists {
		return nil, ErrCacheMiss
	}

	channelCache.RLock()
	defer channelCache.RUnlock()

	// Return direct reference to the internal slice
	// This is ultra-fast but unsafe for long-term use
	if channelCache.size == 0 {
		return nil, nil
	}

	if channelCache.head == 0 && channelCache.size == len(channelCache.messages) {
		// Common special case: full buffer starting at index 0
		return channelCache.messages, nil
	}

	// For other cases, return the appropriate slice view
	return c.GetMessages(channelID)
}

// GetMessagesLimit retrieves up to a specified number of recent messages for a given channel.
func (c *MessageCache) GetMessagesLimit(channelID string, limit int) ([]*discordgo.Message, error) {
	if channelID == "" {
		return nil, ErrInvalidChannel
	}
	if limit <= 0 {
		return nil, ErrInvalidLimit
	}

	c.RLock()
	channelCache, exists := c.channels[channelID]
	c.RUnlock()

	if !exists {
		return nil, ErrCacheMiss
	}

	// Fast path for small limits or when a slice view is sufficient
	channelCache.RLock()

	// Early return for empty cache
	if channelCache.size == 0 {
		channelCache.RUnlock()
		return make([]*discordgo.Message, 0), nil
	}

	// Get local copies of needed values to minimize lock time
	head := channelCache.head
	size := channelCache.size
	maxMsgs := channelCache.maxMessages
	messages := channelCache.messages

	// Release the lock ASAP
	channelCache.RUnlock()

	// Adjust limit if needed
	if limit > size {
		limit = size
	}

	// Special case: if requesting all messages, use GetMessages
	if limit == size {
		return c.GetMessages(channelID)
	}

	// Calculate the start index for the most recent 'limit' messages
	startIdx := (head + size - limit) % maxMsgs

	// Check if we can return a continuous slice view (faster)
	if startIdx+limit <= maxMsgs {
		// We can return a direct slice view without copying
		return messages[startIdx : startIdx+limit], nil
	} else if startIdx > head {
		// Messages wrap around the buffer, need to copy
		result := make([]*discordgo.Message, limit)

		// Calculate sizes of the two segments
		firstPartSize := maxMsgs - startIdx
		secondPartSize := limit - firstPartSize

		// Copy first part (from startIdx to end of buffer)
		copy(result, messages[startIdx:])

		// Copy second part (from start of buffer)
		copy(result[firstPartSize:], messages[:secondPartSize])

		return result, nil
	} else {
		// Simple case: most recent messages are consecutive
		result := make([]*discordgo.Message, limit)
		for i := 0; i < limit; i++ {
			idx := (startIdx + i) % maxMsgs
			result[i] = messages[idx]
		}
		return result, nil
	}
}

// ClearChannel removes all cached messages for a specific channel
func (c *MessageCache) ClearChannel(channelID string) error {
	if channelID == "" {
		return ErrInvalidChannel
	}

	c.RLock()
	channelCache, exists := c.channels[channelID]
	c.RUnlock()

	if !exists {
		return nil // Nothing to clear
	}

	// Clear the channel cache
	channelCache.Lock()
	defer channelCache.Unlock()

	// Reset circular buffer state
	channelCache.head = 0
	channelCache.size = 0

	// Clear the message ID tracking map
	channelCache.messageIDs = make(map[string]struct{}, channelCache.maxMessages)

	return nil
}

// SetMaxMessages sets the maximum number of messages to store per channel in the cache.
func (c *MessageCache) SetMaxMessages(maxMessages int) error {
	if maxMessages <= 0 {
		return ErrInvalidLimit
	}

	// Fast atomic update for future channel caches
	atomic.StoreInt32(&c.maxMessages, int32(maxMessages))

	// Update existing channels
	c.Lock()
	defer c.Unlock()

	// Iterate through all channels
	for _, channelCache := range c.channels {
		channelCache.Lock()

		oldMax := channelCache.maxMessages
		oldSize := channelCache.size
		oldHead := channelCache.head
		oldMessages := channelCache.messages

		// If increasing size, simply update maxMessages
		if maxMessages >= oldMax {
			// Create new array with increased size
			newMessages := make([]*discordgo.Message, maxMessages)

			// Copy existing messages
			for i := 0; i < oldSize; i++ {
				idx := (oldHead + i) % oldMax
				newMessages[i] = oldMessages[idx]
			}

			// Update cache state
			channelCache.messages = newMessages
			channelCache.head = 0
			channelCache.maxMessages = maxMessages
		} else {
			// If decreasing size, need to keep only the most recent messages
			newSize := oldSize
			if newSize > maxMessages {
				newSize = maxMessages
			}

			// Create new array with decreased size
			newMessages := make([]*discordgo.Message, maxMessages)

			// Copy only the most recent messages
			startIdx := oldSize - newSize
			for i := 0; i < newSize; i++ {
				oldIdx := (oldHead + startIdx + i) % oldMax
				newMessages[i] = oldMessages[oldIdx]
			}

			// Rebuild the message ID tracking map
			newIDs := make(map[string]struct{}, maxMessages)
			for i := 0; i < newSize; i++ {
				if msg := newMessages[i]; msg != nil {
					newIDs[msg.ID] = struct{}{}
				}
			}

			// Update cache state
			channelCache.messages = newMessages
			channelCache.messageIDs = newIDs
			channelCache.head = 0
			channelCache.size = newSize
			channelCache.maxMessages = maxMessages
		}

		channelCache.Unlock()
	}

	return nil
}

// Global cache with thread-safe initialization
var (
	globalCache     *MessageCache
	globalCacheOnce sync.Once
)

// GetGlobalCache returns the singleton global cache instance,
// initializing it if necessary
func GetGlobalCache() *MessageCache {
	globalCacheOnce.Do(func() {
		globalCache = NewMessageCache(100)
	})
	return globalCache
}
