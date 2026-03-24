package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// RuleConfig holds per-rule configuration from .archlint.yaml.
type RuleConfig struct {
	Enabled          bool        `yaml:"enabled"`
	ErrorOnViolation bool        `yaml:"error_on_violation"`
	Exclude          []string    `yaml:"exclude"`
	Threshold        interface{} `yaml:"threshold,omitempty"`
}

// RulesConfig is the top-level structure of .archlint.yaml.
type RulesConfig struct {
	Rules map[string]RuleConfig `yaml:"rules"`
}

// LoadRules loads the rules configuration from a .archlint.yaml file.
// Returns nil config (not an error) if the file doesn't exist.
func LoadRules(dir string) (*RulesConfig, error) {
	path := filepath.Join(dir, ".archlint.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg RulesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return &cfg, nil
}

// ExcludesFor returns the exclude patterns for a given rule name.
// Returns nil if the rule is not found or has no excludes.
func (c *RulesConfig) ExcludesFor(rule string) []string {
	if c == nil {
		return nil
	}
	r, ok := c.Rules[rule]
	if !ok {
		return nil
	}
	return r.Exclude
}

// MatchesExclude checks whether a target string matches any of the given
// exclude patterns. Patterns use filepath.Match glob syntax (e.g., "cmd/*",
// "pkg/engine*"). A target matches if the pattern matches the target itself
// or any prefix segment of the target (split by ".").
func MatchesExclude(target string, patterns []string) bool {
	for _, pattern := range patterns {
		// Exact match.
		if matched, _ := filepath.Match(pattern, target); matched {
			return true
		}
		// Match against package prefix (target may be "pkg/storage.SQLitePortfolio").
		if dot := strings.IndexByte(target, '.'); dot >= 0 {
			pkg := target[:dot]
			if matched, _ := filepath.Match(pattern, pkg); matched {
				return true
			}
		}
		// Match the full target against patterns that include a dot
		// (e.g., "pkg/providers/tinkoff.Provider").
		if strings.Contains(pattern, ".") {
			if matched, _ := filepath.Match(pattern, target); matched {
				return true
			}
		}
	}
	return false
}
