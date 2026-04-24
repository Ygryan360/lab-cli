package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Profile struct {
	Name string `json:"name"`
}

type CategoryDefault struct {
	Category    string `json:"category"`
	ProfileName string `json:"profile_name"`
}

type Config struct {
	LabPath          string            `json:"lab_path"`
	DefaultProfile   string            `json:"default_profile"`
	Profiles         []Profile         `json:"profiles"`
	CategoryDefaults []CategoryDefault `json:"category_defaults"`
	Terminal         string            `json:"terminal"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		LabPath:          filepath.Join(home, "lab"),
		DefaultProfile:   "Default",
		Profiles:         []Profile{{Name: "Default"}},
		CategoryDefaults: []CategoryDefault{},
		Terminal:         "kitty",
	}
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lab")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func Load() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			if saveErr := Save(cfg); saveErr != nil {
				return cfg, nil
			}
			return cfg, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0644)
}

func (c *Config) GetProfileForCategory(category string) string {
	for _, cd := range c.CategoryDefaults {
		if cd.Category == category {
			return cd.ProfileName
		}
	}
	return c.DefaultProfile
}

func (c *Config) ProfileExists(name string) bool {
	for _, p := range c.Profiles {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (c *Config) AddProfile(name string) {
	if !c.ProfileExists(name) {
		c.Profiles = append(c.Profiles, Profile{Name: name})
	}
}

func (c *Config) RemoveProfile(name string) {
	var updated []Profile
	for _, p := range c.Profiles {
		if p.Name != name {
			updated = append(updated, p)
		}
	}
	c.Profiles = updated
}

func (c *Config) SetCategoryDefault(category, profileName string) {
	for i, cd := range c.CategoryDefaults {
		if cd.Category == category {
			c.CategoryDefaults[i].ProfileName = profileName
			return
		}
	}
	c.CategoryDefaults = append(c.CategoryDefaults, CategoryDefault{
		Category:    category,
		ProfileName: profileName,
	})
}