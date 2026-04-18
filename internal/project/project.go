package project

import (
	"os"
	"path/filepath"
)

type Project struct {
	Name     string
	Path     string
	Category string
}

type Category struct {
	Name     string
	Projects []Project
}

// Scan reads ~/lab and returns all categories with their projects
func Scan(labPath string) ([]Category, error) {
	entries, err := os.ReadDir(labPath)
	if err != nil {
		return nil, err
	}

	var categories []Category

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		catName := entry.Name()
		catPath := filepath.Join(labPath, catName)

		projects, err := scanCategory(catPath, catName)
		if err != nil {
			continue
		}

		categories = append(categories, Category{
			Name:     catName,
			Projects: projects,
		})
	}

	return categories, nil
}

func scanCategory(catPath, catName string) ([]Project, error) {
	entries, err := os.ReadDir(catPath)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projects = append(projects, Project{
			Name:     entry.Name(),
			Path:     filepath.Join(catPath, entry.Name()),
			Category: catName,
		})
	}
	return projects, nil
}

// Search returns projects matching a query (case-insensitive substring)
func Search(categories []Category, query string) []Project {
	if query == "" {
		return nil
	}
	var results []Project
	q := toLower(query)
	for _, cat := range categories {
		for _, p := range cat.Projects {
			if contains(toLower(p.Name), q) || contains(toLower(p.Category), q) {
				results = append(results, p)
			}
		}
	}
	return results
}

func toLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
