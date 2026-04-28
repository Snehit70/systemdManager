package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	General GeneralConfig       `yaml:"general"`
	UI      UIConfig            `yaml:"ui"`
	Groups  map[string][]string `yaml:"groups"`
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	LogLines        int           `yaml:"log_lines"`
	MaxFollowLines  int           `yaml:"max_follow_lines"`
	Editor          string        `yaml:"editor"`
}

// UIConfig holds UI-related settings.
type UIConfig struct {
	Theme        string `yaml:"theme"`
	ReduceMotion bool   `yaml:"reduce_motion"`
	ShowHidden   bool   `yaml:"show_hidden"`
}

func Default() *Config {
	return &Config{
		General: GeneralConfig{
			RefreshInterval: 2 * time.Second,
			LogLines:        50,
			MaxFollowLines:  1000,
			Editor:          "",
		},
		UI: UIConfig{
			Theme:        "mocha",
			ReduceMotion: false,
			ShowHidden:   false,
		},
		Groups: make(map[string][]string),
	}
}

func (c *GeneralConfig) Validate() {
	if c.RefreshInterval < 1*time.Second {
		c.RefreshInterval = 1 * time.Second
	}
	if c.RefreshInterval > 60*time.Second {
		c.RefreshInterval = 60 * time.Second
	}
	if c.LogLines < 10 {
		c.LogLines = 10
	}
	if c.LogLines > 1000 {
		c.LogLines = 1000
	}
	if c.MaxFollowLines < 100 {
		c.MaxFollowLines = 100
	}
	if c.MaxFollowLines > 10000 {
		c.MaxFollowLines = 10000
	}
}

func Load() (*Config, error) {
	configPath, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		cfg := Default()
		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.General.Validate()

	return cfg, nil
}

func (c *Config) Save() error {
	configPath, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func configPath() (string, error) {
	if envPath := os.Getenv("SYSTEMD_TUI_CONFIG"); envPath != "" {
		return envPath, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	return filepath.Join(configDir, "systemd-tui", "config.yaml"), nil
}

func (c *Config) GetEditor() string {
	if c.General.Editor != "" {
		return c.General.Editor
	}
	if envEditor := os.Getenv("EDITOR"); envEditor != "" {
		return envEditor
	}
	return "vim"
}

func (c *Config) ThemeColors() ThemeColors {
	switch c.UI.Theme {
	case "light":
		return LightTheme
	case "high-contrast":
		return HighContrastTheme
	case "dark":
		return DarkTheme
	default:
		return MochaTheme
	}
}

type ThemeColors struct {
	// Background is left empty when the theme should not paint panel
	// backgrounds — the terminal background shows through. Surface is still
	// used for the modal, status bar, help bar and similar elements that
	// must stand apart from the underlying terminal.
	Background      string
	Surface         string
	SurfaceDeep     string
	Border          string
	BorderActive    string
	Accent          string
	Text            string
	TextMuted       string
	TextDim         string
	StatusActive    string
	StatusFailed    string
	StatusInactive  string
	StatusActivating string
	LogError        string
	LogWarn         string
	LogInfo         string
	LogDebug        string
	SourceUser      string
	SourceSystem    string
	SourceTransient string
	SourceGenerated string
	SourceStatic    string
}

// MochaTheme is a Catppuccin Mocha–inspired palette. Panels stay transparent
// (Background is unset) so the terminal background shows through; only the
// modal, status bar and help bar are filled with Surface/SurfaceDeep.
var MochaTheme = ThemeColors{
	Background:       "",
	Surface:          "#313244", // surface0
	SurfaceDeep:      "#181825", // mantle
	Border:           "#45475a", // surface1
	BorderActive:     "#cba6f7", // mauve
	Accent:           "#89b4fa", // blue
	Text:             "#cdd6f4", // text
	TextMuted:        "#a6adc8", // subtext0
	TextDim:          "#6c7086", // overlay0
	StatusActive:     "#a6e3a1", // green
	StatusFailed:     "#f38ba8", // red
	StatusInactive:   "#6c7086", // overlay0
	StatusActivating: "#fab387", // peach
	LogError:         "#f38ba8",
	LogWarn:          "#f9e2af",
	LogInfo:          "#89b4fa",
	LogDebug:         "#6c7086",
	SourceUser:       "#a6e3a1",
	SourceSystem:     "#89b4fa",
	SourceTransient:  "#f9e2af",
	SourceGenerated:  "#cba6f7",
	SourceStatic:     "#7f849c",
}

// DarkTheme provides the legacy dark color scheme.
var DarkTheme = ThemeColors{
	Background:       "#1A1A2E",
	Surface:          "#16213E",
	SurfaceDeep:      "#0F1A33",
	Border:           "#2D3A5C",
	BorderActive:     "#5B9BF3",
	Accent:           "#5B9BF3",
	Text:             "#E8E8E8",
	TextMuted:        "#9A9A9A",
	TextDim:          "#6B6B6B",
	StatusActive:     "#73F59F",
	StatusFailed:     "#FF6B6B",
	StatusInactive:   "#666666",
	StatusActivating: "#FFA500",
	LogError:         "#FF6B6B",
	LogWarn:          "#FFD93D",
	LogInfo:          "#5B9BF3",
	LogDebug:         "#6B6B6B",
	SourceUser:       "#73F59F",
	SourceSystem:     "#5B9BF3",
	SourceTransient:  "#FFD93D",
	SourceGenerated:  "#C792EA",
	SourceStatic:     "#666666",
}

// LightTheme provides a light color scheme for bright environments.
var LightTheme = ThemeColors{
	Background:       "#FAFAFA",
	Surface:          "#FFFFFF",
	SurfaceDeep:      "#F0F0F0",
	Border:           "#E0E0E0",
	BorderActive:     "#1976D2",
	Accent:           "#1976D2",
	Text:             "#1A1A1A",
	TextMuted:        "#555555",
	TextDim:          "#757575",
	StatusActive:     "#2E7D32",
	StatusFailed:     "#C62828",
	StatusInactive:   "#757575",
	StatusActivating: "#EF6C00",
	LogError:         "#C62828",
	LogWarn:          "#F9A825",
	LogInfo:          "#1976D2",
	LogDebug:         "#757575",
	SourceUser:       "#2E7D32",
	SourceSystem:    "#1976D2",
	SourceTransient: "#F9A825",
	SourceGenerated: "#7B1FA2",
	SourceStatic:    "#757575",
}

// HighContrastTheme provides a high contrast color scheme for accessibility.
var HighContrastTheme = ThemeColors{
	Background:       "#000000",
	Surface:          "#1A1A1A",
	SurfaceDeep:      "#000000",
	Border:           "#FFFFFF",
	BorderActive:     "#00FF00",
	Accent:           "#00FFFF",
	Text:             "#FFFFFF",
	TextMuted:        "#CCCCCC",
	TextDim:          "#AAAAAA",
	StatusActive:     "#00FF00",
	StatusFailed:     "#FF0000",
	StatusInactive:   "#888888",
	StatusActivating: "#FFFF00",
	LogError:         "#FF0000",
	LogWarn:          "#FFFF00",
	LogInfo:          "#00FFFF",
	LogDebug:         "#AAAAAA",
	SourceUser:       "#00FF00",
	SourceSystem:    "#00FFFF",
	SourceTransient: "#FFFF00",
	SourceGenerated: "#FF00FF",
	SourceStatic:    "#888888",
}
