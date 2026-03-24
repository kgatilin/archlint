package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRules(t *testing.T) {
	t.Run("missing file returns nil", func(t *testing.T) {
		cfg, err := LoadRules(t.TempDir())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != nil {
			t.Fatal("expected nil config for missing file")
		}
	})

	t.Run("valid yaml", func(t *testing.T) {
		dir := t.TempDir()
		yaml := `rules:
  coupling:
    enabled: true
    error_on_violation: true
    exclude:
      - cmd/*
      - pkg/engine*
  god_class:
    enabled: true
    exclude: []
`
		if err := os.WriteFile(filepath.Join(dir, ".archlint.yaml"), []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadRules(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if len(cfg.Rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(cfg.Rules))
		}

		excludes := cfg.ExcludesFor("coupling")
		if len(excludes) != 2 {
			t.Fatalf("expected 2 excludes for coupling, got %d", len(excludes))
		}

		excludes = cfg.ExcludesFor("god_class")
		if len(excludes) != 0 {
			t.Fatalf("expected 0 excludes for god_class, got %d", len(excludes))
		}

		excludes = cfg.ExcludesFor("nonexistent")
		if excludes != nil {
			t.Fatalf("expected nil excludes for nonexistent rule, got %v", excludes)
		}
	})
}

func TestMatchesExclude(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		patterns []string
		want     bool
	}{
		{
			name:     "exact match",
			target:   "pkg/models",
			patterns: []string{"pkg/models"},
			want:     true,
		},
		{
			name:     "wildcard match",
			target:   "cmd/main",
			patterns: []string{"cmd/*"},
			want:     true,
		},
		{
			name:     "prefix wildcard",
			target:   "pkg/engine",
			patterns: []string{"pkg/engine*"},
			want:     true,
		},
		{
			name:     "package prefix from dotted target",
			target:   "pkg/storage.SQLitePortfolio",
			patterns: []string{"pkg/storage*"},
			want:     true,
		},
		{
			name:     "specific type exclude",
			target:   "pkg/providers/tinkoff.Provider",
			patterns: []string{"pkg/providers/tinkoff.Provider"},
			want:     true,
		},
		{
			name:     "no match",
			target:   "internal/service",
			patterns: []string{"cmd/*", "pkg/*"},
			want:     false,
		},
		{
			name:     "empty patterns",
			target:   "cmd/main",
			patterns: nil,
			want:     false,
		},
		{
			name:     "method-level target matches package glob",
			target:   "cmd/portfolio-analysis.PortfolioAnalyzer.Analyze",
			patterns: []string{"cmd/*"},
			want:     true,
		},
		{
			name:     "method-level exact match",
			target:   "cmd/portfolio-analysis.PortfolioAnalyzer.Analyze",
			patterns: []string{"cmd/portfolio-analysis.PortfolioAnalyzer.Analyze"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesExclude(tt.target, tt.patterns)
			if got != tt.want {
				t.Errorf("MatchesExclude(%q, %v) = %v, want %v", tt.target, tt.patterns, got, tt.want)
			}
		})
	}
}
