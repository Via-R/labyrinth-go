package core

import (
	"fmt"
	"os"

	toml "github.com/BurntSushi/toml"
)

// configuration structure
type (
	builder struct {
		Complexity              float64 `toml:"complexity"`
		MaxAreaToCoverWithWalls float64 `toml:"max_area_to_cover_with_walls"`
		OnlyOnePathNearFinish   bool    `toml:"only_one_path_near_finish"`
		LabyrinthBuilderAtempts uint    `toml:"labyrinth_builder_atempts"`
	}
	configuration struct {
		Builder builder `toml:"builder"`
	}
)

// Custom error for Configuration
func (configuration) Error(s string) error {
	return fmt.Errorf("Configuration error: %v", s)
}

// Validate configuration values
func (c *configuration) validate() error {
	if c.Builder.Complexity < 0 || c.Builder.Complexity > 100 {
		return c.Error("Complexity (percentage) cannot be less than 0 or over 100")
	}
	if c.Builder.MaxAreaToCoverWithWalls <= 0 || c.Builder.MaxAreaToCoverWithWalls > 100 {
		return c.Error("Max area to cover with walls (percentage) cannot be less or equal to 0 or over 1")
	}

	return nil
}

// Load configuration values from TOML file under 'filename'
func (c *configuration) LoadFromFile(filename string) error {
	if blob, err := os.ReadFile(filename); err != nil {
		return err
	} else if _, err := toml.Decode(string(blob), c); err != nil {
		return err
	}

	if err := c.validate(); err != nil {
		return err
	}

	return nil
}
