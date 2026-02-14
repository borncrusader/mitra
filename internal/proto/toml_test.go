package proto

import (
	"bytes"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestProtoToTOML(t *testing.T) {
	cfg := &Config{
		Server: &ServerConfig{
			Port: ":9999",
		},
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)

	if err := encoder.Encode(cfg); err != nil {
		t.Fatalf("Failed to encode proto to TOML: %v", err)
	}

	output := buf.String()
	t.Logf("TOML output:\n%s", output)

	if output == "" {
		t.Error("TOML output is empty")
	}
}
