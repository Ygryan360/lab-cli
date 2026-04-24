package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ygryan360/lab-cli/internal/config"
	"github.com/ygryan360/lab-cli/internal/history"
	"github.com/ygryan360/lab-cli/internal/project"
	"github.com/ygryan360/lab-cli/internal/terminal"
	"github.com/ygryan360/lab-cli/internal/tui"
)

const version = "1.0.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		runTUI()
		return
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "open":
		runOpen(rest)
	case "term":
		runTerm(rest)
	case "recent":
		runRecent()
	case "search":
		runSearch(rest)
	case "profiles":
		runProfiles(rest)
	case "config":
		runConfig(rest)
	case "version", "--version", "-v":
		fmt.Printf("lab v%s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		// Try to open as project name directly
		runOpen(args)
	}
}

func loadAll() (*config.Config, *history.History, []project.Category) {
	cfg, err := config.Load()
	if err != nil {
		die("Failed to load config: %v", err)
	}

	hist, err := history.Load()
	if err != nil {
		die("Failed to load history: %v", err)
	}

	cats, err := project.Scan(cfg.LabPath)
	if err != nil {
		die("Failed to scan lab directory (%s): %v", cfg.LabPath, err)
	}

	return cfg, hist, cats
}

func runTUI() {
	cfg, hist, cats := loadAll()
	app := tui.NewApp(cfg, hist, cats)
	if err := app.Run(); err != nil {
		die("TUI error: %v", err)
	}
}

func runTerm(args []string) {
	if len(args) == 0 {
		die("Usage: lab term <project-name>")
	}
	name := args[0]
	cfg, _, cats := loadAll()

	var found *project.Project
	for _, cat := range cats {
		for _, p := range cat.Projects {
			if p.Name == name {
				cp := p
				found = &cp
				break
			}
		}
		if found != nil {
			break
		}
	}
	if found == nil {
		die("Project %q not found", name)
	}

	term := cfg.Terminal
	if term == "" {
		term = terminal.Detect()
	}

	if err := terminal.OpenTab(term, found.Path, found.Name); err != nil {
		die("%v", err)
	}
	fmt.Printf("⌨  Opened terminal tab → %s\n", found.Name)
}

func runOpen(args []string) {
	if len(args) == 0 {
		die("Usage: lab open <project-name> [--profile <name>]")
	}
	name := args[0]
	profile := ""
	for i, a := range args {
		if (a == "--profile" || a == "-p") && i+1 < len(args) {
			profile = args[i+1]
		}
	}

	cfg, hist, cats := loadAll()
	if err := tui.OpenDirect(cfg, hist, cats, name, profile); err != nil {
		die("%v", err)
	}
}

func runRecent() {
	_, hist, _ := loadAll()
	recent := hist.Recent(15)

	if len(recent) == 0 {
		fmt.Println("No recent projects.")
		return
	}

	fmt.Printf("\n%-22s %-14s %-12s %-16s %s\n",
		"PROJECT", "CATEGORY", "LAST OPENED", "PROFILE", "OPENS")
	fmt.Println(strings.Repeat("─", 80))

	for _, e := range recent {
		fmt.Printf("%-22s %-14s %-12s %-16s %d\n",
			truncate(e.Name, 21),
			truncate(e.Category, 13),
			history.FormatTimeAgo(e.LastOpenedAt),
			truncate(e.Profile, 15),
			e.OpenCount,
		)
	}
	fmt.Println()
}

func runSearch(args []string) {
	if len(args) == 0 {
		die("Usage: lab search <query>")
	}
	query := strings.Join(args, " ")
	_, _, cats := loadAll()

	results := project.Search(cats, query)
	if len(results) == 0 {
		fmt.Printf("No projects found matching %q\n", query)
		return
	}

	fmt.Printf("\n%-22s %s\n", "PROJECT", "CATEGORY")
	fmt.Println(strings.Repeat("─", 40))
	for _, p := range results {
		fmt.Printf("%-22s %s\n", p.Name, p.Category)
	}
	fmt.Println()
}

func runProfiles(args []string) {
	cfg, _, _ := loadAll()

	if len(args) == 0 {
		// List profiles
		fmt.Println("\nConfigured VSCode profiles:")
		fmt.Println(strings.Repeat("─", 40))
		for _, p := range cfg.Profiles {
			marker := "  "
			if p.Name == cfg.DefaultProfile {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, p.Name)
		}
		fmt.Printf("\n* = default profile\n")
		fmt.Println("\nCategory defaults:")
		if len(cfg.CategoryDefaults) == 0 {
			fmt.Println("  (none)")
		}
		for _, cd := range cfg.CategoryDefaults {
			fmt.Printf("  %-16s → %s\n", cd.Category, cd.ProfileName)
		}
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  lab profiles add <name>")
		fmt.Println("  lab profiles remove <name>")
		fmt.Println("  lab profiles default <name>")
		fmt.Println("  lab profiles set-category <category> <profile>")
		return
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "add":
		if len(rest) == 0 {
			die("Usage: lab profiles add <name>")
		}
		name := strings.Join(rest, " ")
		cfg.AddProfile(name)
		config.Save(cfg)
		fmt.Printf("✓ Profile %q added\n", name)

	case "remove", "rm":
		if len(rest) == 0 {
			die("Usage: lab profiles remove <name>")
		}
		name := strings.Join(rest, " ")
		if name == cfg.DefaultProfile {
			die("Cannot remove the default profile")
		}
		cfg.RemoveProfile(name)
		config.Save(cfg)
		fmt.Printf("✓ Profile %q removed\n", name)

	case "default":
		if len(rest) == 0 {
			die("Usage: lab profiles default <name>")
		}
		name := strings.Join(rest, " ")
		if !cfg.ProfileExists(name) {
			die("Profile %q does not exist", name)
		}
		cfg.DefaultProfile = name
		config.Save(cfg)
		fmt.Printf("✓ Default profile set to %q\n", name)

	case "set-category":
		if len(rest) < 2 {
			die("Usage: lab profiles set-category <category> <profile>")
		}
		category := rest[0]
		profileName := strings.Join(rest[1:], " ")
		if !cfg.ProfileExists(profileName) {
			die("Profile %q does not exist. Add it first with: lab profiles add %q", profileName, profileName)
		}
		cfg.SetCategoryDefault(category, profileName)
		config.Save(cfg)
		fmt.Printf("✓ Category %q will use profile %q by default\n", category, profileName)

	default:
		die("Unknown profiles subcommand: %s", sub)
	}
}

func runConfig(args []string) {
	cfg, _, _ := loadAll()

	if len(args) == 0 {
		fmt.Printf("\nlab configuration\n")
		fmt.Println(strings.Repeat("─", 40))
		fmt.Printf("Lab path:        %s\n", cfg.LabPath)
		fmt.Printf("Default profile: %s\n", cfg.DefaultProfile)
		fmt.Printf("Config file:     %s\n", config.ConfigPath())
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  lab config set-path <path>")
		return
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "set-path":
		if len(rest) == 0 {
			die("Usage: lab config set-path <path>")
		}
		cfg.LabPath = rest[0]
		config.Save(cfg)
		fmt.Printf("✓ Lab path set to %q\n", rest[0])
	default:
		die("Unknown config subcommand: %s", sub)
	}
}

func printHelp() {
	fmt.Printf(`
lab v%s — Project Launcher

USAGE:
  lab                          Open interactive TUI
  lab open <project>           Open a project in VSCode
  lab open <project> -p <prof> Open with specific profile
  lab term <project>           Open a terminal tab in project directory
  lab recent                   List recently opened projects
  lab search <query>           Search projects
  lab profiles                 Manage VSCode profiles
  lab config                   Show/edit configuration
  lab version                  Show version

EXAMPLES:
  lab
  lab open univarchives
  lab open brainbounty --profile "React/Frontend"
  lab profiles add "Laravel/PHP"
  lab profiles set-category sites "Laravel/PHP"
  lab profiles default "Laravel/PHP"

CONFIG: ~/.config/lab/config.json
`, version)
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "lab: "+format+"\n", args...)
	os.Exit(1)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-2] + ".."
}