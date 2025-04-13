// This file provides backward compatibility for existing code
package dgocacheler

import "github.com/bwmarrin/discordgo"

// Cache is maintained for backward compatibility with existing code.
// New code should use GetGlobalCache() instead.
//
// Deprecated: This interface is deprecated and will be removed in future versions.
var Cache interface {
	// Deprecated.
	//
	// Use this instead of Cache directly:
	//      cache := dgocachler.GetGlobalCache()
	//      cache.AddMessage()
	AddMessage(channelID string, message *discordgo.Message) error
	// Deprecated.
	//
	// Use this instead of Cache directly:
	//      cache := dgocachler.GetGlobalCache()
	//      cache.AddMessages()
	AddMessages(channelID string, messages []*discordgo.Message) error
	// Deprecated.
	//
	// Use this instead of Cache directly:
	//      cache := dgocachler.GetGlobalCache()
	//      cache.GetMessages()
	GetMessages(channelID string) ([]*discordgo.Message, error)
	// Deprecated.
	//
	// Use this instead of Cache directly:
	//      cache := dgocachler.GetGlobalCache()
	//      cache.GetMessagesLimit()
	GetMessagesLimit(channelID string, limit int) ([]*discordgo.Message, error)
	// Deprecated.
	//
	// Use this instead of Cache directly:
	//      cache := dgocachler.GetGlobalCache()
	//      cache.ClearChannel()
	ClearChannel(channelID string) error
	// Deprecated.
	//
	// Use this instead of Cache directly:
	//      cache := dgocachler.GetGlobalCache()
	//      cache.SetMaxMessages()
	SetMaxMessages(maxMessages int) error
}

func init() {
	// Initialize the Cache variable with the global cache instance
	Cache = GetGlobalCache()
}
