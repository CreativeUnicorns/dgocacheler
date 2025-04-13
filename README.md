[![go](https://github.com/CreativeUnicorns/dgocacheler/actions/workflows/go.yml/badge.svg)](https://github.com/CreativeUnicorns/dgocacheler/actions/workflows/go.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/CreativeUnicorns/dgocacheler.svg)](https://pkg.go.dev/github.com/CreativeUnicorns/dgocacheler)
# dgocacheler

A high-performance, concurrency-safe cache for storing Discord messages by channel, designed for seamless integration with the `discordgo` library.

## Features

- **Thread-safe operations**: Fine-grained channel-level locking for maximum concurrency
- **High performance**: Optimized for minimal overhead in both single and multi-threaded scenarios
- **Memory efficient**: Smart memory management with minimal allocations
- **Duplicate prevention**: Automatic detection and handling of duplicate messages
- **Flexible API**: Both safe and ultra-fast access methods for different use cases
- **Backward compatible**: Smooth transition from previous versions

## Performance

Benchmarks show significant improvements over previous versions:

- 46% faster for concurrent read/write operations
- 391% faster for multi-channel operations
- 99% reduction in memory allocations for multi-channel scenarios
- Zero-allocation read operations

## Installation

```bash
go get github.com/CreativeUnicorns/dgocacheler@v1.0.0
```

## Usage

### Basic Usage

```go
// Get the global cache
cache := dgocacheler.GetGlobalCache()

// Add a message to the cache
err := cache.AddMessage("channel123", message)
if err != nil {
    log.Printf("Failed to add message: %v", err)
}

// Add multiple messages at once
err = cache.AddMessages("channel123", messages)
if err != nil {
    log.Printf("Failed to add messages: %v", err)
}

// Retrieve messages
messages, err := cache.GetMessages("channel123")
if err != nil {
    if errors.Is(err, dgocacheler.ErrCacheMiss) {
        log.Printf("No messages in channel")
    } else {
        log.Printf("Error retrieving messages: %v", err)
    }
}

// Retrieve limited number of most recent messages
messages, err = cache.GetMessagesLimit("channel123", 10)
if err != nil {
    log.Printf("Error retrieving messages: %v", err)
}

// Set maximum messages per channel
err = cache.SetMaxMessages(500)
if err != nil {
    log.Printf("Error setting max messages: %v", err)
}

// Clear a channel's messages
err = cache.ClearChannel("channel123")
if err != nil {
    log.Printf("Error clearing channel: %v", err)
}
```

### Performance-Critical Usage

For performance-critical code paths:

```go
// Ultra-fast message retrieval (unsafe for long-term reference)
messages, err := cache.GetMessagesUnsafe("channel123")
if err != nil {
    log.Printf("Error retrieving messages: %v", err)
}
```

### Backward Compatibility

Code written for previous versions continues to work:

```go
// Legacy code still works
dgocacheler.Cache.AddMessage("channel123", message)

// But new code should use
dgocacheler.GetGlobalCache().AddMessage("channel123", message)
```

For more examples, refer to the `examples/` directory.

## Implementation Details

- **True circular buffer**: Efficient O(1) operations for all common operations
- **Optimized locking**: Minimized lock scopes to reduce contention
- **Atomic operations**: Lock-free access to configuration values
- **Smart slice handling**: Direct slice referencing when possible

## Contributing

Contributions are welcome! Please feel free to submit a pull request.

## License

Distributed under the MIT License. See `LICENSE` file for more information.