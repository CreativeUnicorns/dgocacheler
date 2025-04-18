package subpkg

import (
	"fmt"

	"github.com/CreativeUnicorns/dgocacheler"
	"github.com/bwmarrin/discordgo"
)

func AppendChannel2() {
	// Add some messages to the cache
	for i := 0; i < 20; i++ {
		msg := &discordgo.Message{ID: fmt.Sprintf("%d", i), Content: fmt.Sprintf("Ch 2. Message %d", i)}
		// Use GetGlobalCache() instead of accessing Cache directly
		err := dgocacheler.GetGlobalCache().AddMessage("channel2", msg)
		if err != nil {
			fmt.Printf("Error adding message: %v\n", err)
		}
	}
}
