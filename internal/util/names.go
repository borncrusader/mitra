package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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

var greekNames = []string{
	"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta",
	"iota", "kappa", "lambda", "mu", "nu", "xi", "omicron", "pi",
	"rho", "sigma", "tau", "upsilon", "phi", "chi", "psi", "omega",
}

func NextGreekName(existingBranches []string, prefix string) string {
	used := make(map[string]bool)
	for _, b := range existingBranches {
		used[strings.TrimPrefix(b, prefix)] = true
	}
	for _, name := range greekNames {
		if !used[name] {
			return name
		}
	}
	return RandomName()
}

func RandomName() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	timestamp := time.Now().Unix()
	hexTimestamp := fmt.Sprintf("%x", timestamp)

	return fmt.Sprintf("%s-%s-%s", adj, noun, hexTimestamp)
}

// ExtractTimestamp extracts the Unix timestamp from a generated name
// Returns the timestamp in seconds, or 0 if the name format is invalid
func ExtractTimestamp(name string) int64 {
	parts := strings.Split(name, "-")
	if len(parts) != 3 {
		return 0
	}

	// Last part is the hex timestamp
	hexTimestamp := parts[2]
	timestamp, err := strconv.ParseInt(hexTimestamp, 16, 64)
	if err != nil {
		return 0
	}

	return timestamp
}
