// Package cacheler provides a concurrency-safe cache designed specifically
// for storing Discord messages by channel. It integrates smoothly with the discordgo package.
//
// The dgocacheler package ensures that all operations are safe to use concurrently
// and manages memory efficiently by enforcing a maximum number of messages per channel.
//
// The package is designed to be simple to use and easy to integrate with existing chatbot handlers code.
// The dgocacheler package also provides a global `Cache` that can be used across multiple packages to help avoid circular dependencies.
package dgocacheler
