package util

import (
	"fmt"
	"math/rand"
	"time"
)

var adjectives = []string{
	"happy", "clever", "brave", "calm", "bright",
	"swift", "gentle", "quiet", "wise", "kind",
	"bold", "quick", "warm", "cool", "sharp",
	"smooth", "steady", "strong", "light", "dark",
}

var nouns = []string{
	"river", "mountain", "forest", "ocean", "valley",
	"cloud", "storm", "breeze", "thunder", "lightning",
	"star", "moon", "sun", "sky", "earth",
	"fire", "water", "wind", "stone", "tree",
}

func RandomName() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	timestamp := time.Now().Unix()
	hexTimestamp := fmt.Sprintf("%x", timestamp)

	return fmt.Sprintf("%s-%s-%s", adj, noun, hexTimestamp)
}
