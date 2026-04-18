# lab CLI — Project Launcher CLI

A fast terminal-based project launcher for `~/lab`, built in Go with zero external dependencies.

## Features

- **Interactive TUI** — navigate projects by category, open with a keypress
- **VSCode profile management** — profile saved per-project, asked only on first open
- **Profile priority** : `project-specific > category default > global default`  
- **Recent projects** — with metadata (open count, last opened, profile used)
- **Fuzzy search** — across all projects and categories
- **CLI mode** — scriptable commands for power users

## Installation

```bash
# Clone / copy the project, then:
chmod +x install.sh
./install.sh

# Or manually:
make install
```

## Usage

```bash
lab                              # Open TUI
lab open project            # Open directly
lab open project -p "PHP"   # Open with specific profile
lab recent                       # List recent projects
lab search laravel               # Search projects
lab profiles                     # List profiles
lab profiles add "Laravel/PHP"
lab profiles remove "Node/Backend"
lab profiles default "Laravel/PHP"
lab profiles set-category sites "Laravel/PHP"
lab config                       # Show config
lab config set-path ~/lab
lab help
```

## TUI Keybindings

| Key   | Action |
|-------|--------|
| `↑↓`  | Navigate |
| `Tab` | Switch panel (Categories / Projects / Recent) |
| `Enter` | Open selected project |
| `/` | Search |
| `p` | Change profile for selected project |
| `r` | Jump to Recent panel |
| `q` | Quit |

## Config files

```
~/.config/lab/
├── config.json    # profiles, defaults, lab path
└── history.json   # recent projects + metadata
```

## Project structure expected

```
~/lab
├── __sandbox/
│   └── my-test-project/
├── expo__apk/
│   └── my-app/
├── sites/
│   └── project/
└── tools/
    └── my-script/
```

Each subdirectory of `~/lab` is treated as a **category**.  
Each subdirectory inside a category is treated as a **project**.
