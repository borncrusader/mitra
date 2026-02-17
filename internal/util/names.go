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
	"noble", "silver", "golden", "crystal", "ancient",
	"mystic", "cosmic", "azure", "crimson", "emerald",
	"tranquil", "vibrant", "serene", "radiant", "lucid",
	"nimble", "elegant", "graceful", "fierce", "resilient",
	"amber", "violet", "scarlet", "jade", "pearl",
	"silent", "mighty", "infinite", "ethereal", "pristine",
}

var nouns = []string{
	"river", "mountain", "forest", "ocean", "valley",
	"cloud", "storm", "breeze", "thunder", "lightning",
	"star", "moon", "sun", "sky", "earth",
	"fire", "water", "wind", "stone", "tree",
	"phoenix", "dragon", "wolf", "eagle", "falcon",
	"glacier", "canyon", "meadow", "aurora", "comet",
	"cascade", "horizon", "twilight", "dawn", "dusk",
	"summit", "reef", "tide", "wave", "brook",
	"nebula", "galaxy", "cosmos", "quasar", "prism",
	"shadow", "echo", "whisper", "spark", "ember",
}

func RandomName() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	timestamp := time.Now().Unix()
	hexTimestamp := fmt.Sprintf("%x", timestamp)

	return fmt.Sprintf("%s-%s-%s", adj, noun, hexTimestamp)
}
