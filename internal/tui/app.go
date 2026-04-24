package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/ygryan360/lab-cli/internal/config"
	"github.com/ygryan360/lab-cli/internal/history"
	"github.com/ygryan360/lab-cli/internal/project"
	"github.com/ygryan360/lab-cli/internal/terminal"
	"github.com/ygryan360/lab-cli/internal/vscode"
)

type panel int

const (
	panelCategories panel = iota
	panelProjects
	panelRecent
)

type mode int

const (
	modeNormal mode = iota
	modeSearch
	modeProfilePick
	modeConfirmOpen
)

type App struct {
	cfg        *config.Config
	hist       *history.History
	categories []project.Category
	recent     []history.Entry

	activePanel  panel
	catCursor    int
	projCursor   int
	recentCursor int

	searchQuery   string
	searchResults []project.Project

	currentMode   mode
	profileCursor int

	pendingProject *project.Project
	pendingProfile string

	width  int
	height int

	statusMsg string
}

func NewApp(cfg *config.Config, hist *history.History, cats []project.Category) *App {
	return &App{
		cfg:        cfg,
		hist:       hist,
		categories: cats,
		recent:     hist.Recent(10),
		width:      80,
		height:     24,
	}
}

func (a *App) Run() error {
	if err := EnableRawMode(); err != nil {
		return err
	}
	defer DisableRawMode()

	fmt.Print(CursorHide)
	fmt.Print(ClearScreen)
	defer fmt.Print(CursorShow)
	defer fmt.Print(ClearScreen + CursorHome)

	for {
		a.width, a.height = TermSize()
		a.render()

		key := ReadKey()
		if quit := a.handleKey(key); quit {
			return nil
		}
	}
}

func (a *App) handleKey(k KeyEvent) bool {
	switch a.currentMode {
	case modeSearch:
		return a.handleSearchKey(k)
	case modeProfilePick:
		return a.handleProfilePickKey(k)
	case modeConfirmOpen:
		return a.handleConfirmKey(k)
	default:
		return a.handleNormalKey(k)
	}
}

func (a *App) handleNormalKey(k KeyEvent) bool {
	switch k.Type {
	case KeyCtrlC, KeyCtrlD:
		return true
	case KeyRune:
		switch k.Ch {
		case 'q', 'Q':
			return true
		case '/':
			a.currentMode = modeSearch
			a.searchQuery = ""
			a.searchResults = nil
		case 'p':
			p := a.selectedProject()
			if p != nil {
				a.pendingProject = p
				a.profileCursor = 0
				a.currentMode = modeProfilePick
			}
		case 't':
			p := a.selectedProject()
			if p != nil {
				a.openTerminalTab(p)
			}
		case 'r':
			a.activePanel = panelRecent
		case '1':
			a.activePanel = panelCategories
		case '2':
			a.activePanel = panelProjects
		}
	case KeyTab:
		a.cyclePanel()
	case KeyUp:
		a.moveCursorUp()
	case KeyDown:
		a.moveCursorDown()
	case KeyEnter:
		p := a.selectedProject()
		if p != nil {
			a.openProject(p)
		}
	case KeyEsc:
		a.statusMsg = ""
	}
	return false
}

func (a *App) handleSearchKey(k KeyEvent) bool {
	switch k.Type {
	case KeyEsc, KeyCtrlC:
		a.currentMode = modeNormal
		a.searchQuery = ""
		a.searchResults = nil
	case KeyEnter:
		if len(a.searchResults) > 0 {
			p := &a.searchResults[0]
			a.openProject(p)
			a.currentMode = modeNormal
			a.searchQuery = ""
			a.searchResults = nil
		}
	case KeyBackspace:
		if len(a.searchQuery) > 0 {
			a.searchQuery = a.searchQuery[:len(a.searchQuery)-1]
		}
		a.searchResults = project.Search(a.categories, a.searchQuery)
	case KeyRune:
		a.searchQuery += string(k.Ch)
		a.searchResults = project.Search(a.categories, a.searchQuery)
	}
	return false
}

func (a *App) handleProfilePickKey(k KeyEvent) bool {
	profiles := a.cfg.Profiles
	switch k.Type {
	case KeyEsc:
		a.currentMode = modeNormal
	case KeyUp:
		if a.profileCursor > 0 {
			a.profileCursor--
		}
	case KeyDown:
		if a.profileCursor < len(profiles)-1 {
			a.profileCursor++
		}
	case KeyEnter:
		if len(profiles) > 0 {
			chosen := profiles[a.profileCursor].Name
			if a.pendingProject != nil {
				a.launchVSCode(a.pendingProject, chosen)
			}
			a.currentMode = modeNormal
		}
	}
	return false
}

func (a *App) handleConfirmKey(k KeyEvent) bool {
	switch k.Type {
	case KeyEnter:
		if a.pendingProject != nil {
			a.launchVSCode(a.pendingProject, a.pendingProfile)
		}
		a.currentMode = modeNormal
	case KeyEsc:
		a.currentMode = modeNormal
	}
	return false
}

func (a *App) cyclePanel() {
	switch a.activePanel {
	case panelCategories:
		a.activePanel = panelProjects
	case panelProjects:
		a.activePanel = panelRecent
	case panelRecent:
		a.activePanel = panelCategories
	}
}

func (a *App) moveCursorUp() {
	switch a.activePanel {
	case panelCategories:
		if a.catCursor > 0 {
			a.catCursor--
			a.projCursor = 0
		}
	case panelProjects:
		if a.projCursor > 0 {
			a.projCursor--
		}
	case panelRecent:
		if a.recentCursor > 0 {
			a.recentCursor--
		}
	}
}

func (a *App) moveCursorDown() {
	switch a.activePanel {
	case panelCategories:
		if a.catCursor < len(a.categories)-1 {
			a.catCursor++
			a.projCursor = 0
		}
	case panelProjects:
		projs := a.currentProjects()
		if a.projCursor < len(projs)-1 {
			a.projCursor++
		}
	case panelRecent:
		if a.recentCursor < len(a.recent)-1 {
			a.recentCursor++
		}
	}
}

func (a *App) currentProjects() []project.Project {
	if len(a.categories) == 0 {
		return nil
	}
	if a.catCursor >= len(a.categories) {
		return nil
	}
	return a.categories[a.catCursor].Projects
}

func (a *App) selectedProject() *project.Project {
	switch a.activePanel {
	case panelProjects:
		projs := a.currentProjects()
		if len(projs) == 0 || a.projCursor >= len(projs) {
			return nil
		}
		p := projs[a.projCursor]
		return &p
	case panelRecent:
		if len(a.recent) == 0 || a.recentCursor >= len(a.recent) {
			return nil
		}
		e := a.recent[a.recentCursor]
		return &project.Project{Name: e.Name, Path: e.Path, Category: e.Category}
	default:
		return nil
	}
}

func (a *App) openProject(p *project.Project) {
	// Check if project has a saved profile
	if savedProfile, ok := a.hist.GetProfile(p.Path); ok {
		a.launchVSCode(p, savedProfile)
		return
	}

	// Check category default
	catDefault := a.cfg.GetProfileForCategory(p.Category)
	if catDefault != "" && catDefault != a.cfg.DefaultProfile {
		a.launchVSCode(p, catDefault)
		return
	}

	// First time: if only one profile (Default), open directly
	if len(a.cfg.Profiles) <= 1 {
		a.launchVSCode(p, a.cfg.DefaultProfile)
		return
	}

	// Otherwise ask
	a.pendingProject = p
	a.profileCursor = 0
	a.currentMode = modeProfilePick
}

func (a *App) openTerminalTab(p *project.Project) {
	term := a.cfg.Terminal
	if term == "" {
		term = terminal.Detect()
	}
	if err := terminal.OpenTab(term, p.Path, p.Name); err != nil {
		a.statusMsg = fmt.Sprintf("⚠  Terminal error: %v", err)
		return
	}
	a.statusMsg = fmt.Sprintf("⌨  Opened tab → %s", p.Name)
}

func (a *App) launchVSCode(p *project.Project, profile string) {
	if !vscode.IsInstalled() {
		a.statusMsg = "⚠  VSCode (code) not found in PATH"
		return
	}
	if err := vscode.Open(p.Path, profile); err != nil {
		a.statusMsg = fmt.Sprintf("⚠  Error: %v", err)
		return
	}

	a.hist.Record(p.Name, p.Path, p.Category, profile)
	_ = history.Save(a.hist)
	a.recent = a.hist.Recent(10)
	a.statusMsg = fmt.Sprintf("✓  Opened %s [%s]", p.Name, profile)
}

// ──────────────────────────────────────────────
// Rendering — cursor-positioned, no string padding with ANSI codes
// ──────────────────────────────────────────────

// writeLine writes a plain line at current cursor position then moves down
func writeLine(sb *strings.Builder, row, col int, s string) {
	sb.WriteString(MoveCursor(row, col))
	sb.WriteString(s)
}

// plainLen returns the visual length of a string (ignoring ANSI escape codes)
func plainLen(s string) int {
	n := 0
	inEsc := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			inEsc = true
		} else if inEsc && s[i] == 'm' {
			inEsc = false
		} else if !inEsc {
			n++
		}
	}
	return n
}

// padTo pads a (possibly ANSI-colored) string to a given visual width
func padTo(s string, width int) string {
	n := plainLen(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func (a *App) render() {
	var sb strings.Builder
	sb.WriteString(ClearScreen)
	sb.WriteString(CursorHome)

	row := 1
	row = a.renderHeader(&sb, row)
	row = a.renderMainPanels(&sb, row)
	row = a.renderRecent(&sb, row)
	a.renderFooter(&sb, row)

	if a.currentMode == modeProfilePick {
		a.renderProfilePicker(&sb)
	}

	fmt.Print(sb.String())
}

func (a *App) renderHeader(sb *strings.Builder, row int) int {
	title := Bold + FgCyan + " LAB" + Reset + Dim + " — Project Launcher" + Reset
	ver := Dim + "v1.0.0" + Reset

	writeLine(sb, row, 1, title)
	// version flush right
	vCol := a.width - 6
	if vCol < 25 {
		vCol = 25
	}
	sb.WriteString(MoveCursor(row, vCol) + ver)
	row++
	writeLine(sb, row, 1, Dim+strings.Repeat("─", a.width)+Reset)
	return row + 1
}

func (a *App) renderMainPanels(sb *strings.Builder, startRow int) int {
	catW := 24   // visual width of category column
	projCol := catW + 3 // start column for projects (1-based)

	// Titles row
	catTitle := a.panelTitle("CATEGORIES", a.activePanel == panelCategories)
	projTitle := a.panelTitle("PROJECTS", a.activePanel == panelProjects)
	writeLine(sb, startRow, 2, catTitle)
	writeLine(sb, startRow, projCol, projTitle)

	// Divider row
	writeLine(sb, startRow+1, 2, Dim+strings.Repeat("─", catW)+Reset)
	writeLine(sb, startRow+1, projCol, Dim+strings.Repeat("─", a.width-projCol)+Reset)

	panelHeight := 10
	projs := a.currentProjects()

	catOffset := 0
	if a.catCursor >= panelHeight {
		catOffset = a.catCursor - panelHeight + 1
	}
	projOffset := 0
	if a.projCursor >= panelHeight {
		projOffset = a.projCursor - panelHeight + 1
	}

	for i := 0; i < panelHeight; i++ {
		lineRow := startRow + 2 + i

		// — Category cell —
		catIdx := i + catOffset
		// clear the cell first
		writeLine(sb, lineRow, 1, strings.Repeat(" ", catW+2))

		if catIdx < len(a.categories) {
			cat := a.categories[catIdx]
			count := fmt.Sprintf("(%d)", len(cat.Projects))
			name := Truncate(cat.Name, catW-len(count)-2)
			plain := name + " " + count

			if catIdx == a.catCursor {
				if a.activePanel == panelCategories {
					writeLine(sb, lineRow, 1, BgCyan+FgBlack+padTo(" > "+plain, catW+2)+Reset)
				} else {
					writeLine(sb, lineRow, 1, FgCyan+" > "+plain+Reset)
				}
			} else {
				writeLine(sb, lineRow, 1, Dim+"   "+plain+Reset)
			}
		}

		// — Project cell —
		projIdx := i + projOffset
		// clear the cell
		projCellW := a.width - projCol
		if projCellW < 1 {
			projCellW = 1
		}
		writeLine(sb, lineRow, projCol, strings.Repeat(" ", projCellW))

		if projIdx < len(projs) {
			proj := projs[projIdx]
			hasSaved := ""
			if _, ok := a.hist.GetProfile(proj.Path); ok {
				hasSaved = " ·"
			}
			maxName := projCellW - 5
			if maxName < 1 {
				maxName = 1
			}
			name := Truncate(proj.Name, maxName)

			if projIdx == a.projCursor {
				if a.activePanel == panelProjects {
					writeLine(sb, lineRow, projCol,
						BgCyan+FgBlack+padTo(" > "+name+hasSaved, projCellW)+Reset)
				} else {
					writeLine(sb, lineRow, projCol, FgCyan+" > "+name+Reset+Dim+hasSaved+Reset)
				}
			} else {
				writeLine(sb, lineRow, projCol, Dim+"   "+name+hasSaved+Reset)
			}
		}
	}

	return startRow + 2 + panelHeight + 1
}

func (a *App) renderRecent(sb *strings.Builder, startRow int) int {
	title := a.panelTitle("RECENTLY OPENED", a.activePanel == panelRecent)
	writeLine(sb, startRow, 2, title)
	writeLine(sb, startRow+1, 1, Dim+strings.Repeat("─", a.width)+Reset)

	if len(a.recent) == 0 {
		writeLine(sb, startRow+2, 3, Dim+"No recent projects"+Reset)
		return startRow + 4
	}

	maxRows := 5
	if maxRows > len(a.recent) {
		maxRows = len(a.recent)
	}

	// Column positions (1-based)
	colName    := 3
	colCat     := 26
	colTime    := 42
	colProfile := 56
	colCount   := a.width - 5

	for i := 0; i < maxRows; i++ {
		e := a.recent[i]
		lineRow := startRow + 2 + i

		// Clear row
		writeLine(sb, lineRow, 1, strings.Repeat(" ", a.width))

		timeStr    := history.FormatTimeAgo(e.LastOpenedAt)
		countStr   := fmt.Sprintf("%dx", e.OpenCount)
		nameStr    := Truncate(e.Name, colCat-colName-1)
		catStr     := Truncate(e.Category, colTime-colCat-1)
		profileStr := Truncate(e.Profile, colCount-colProfile-1)

		if i == a.recentCursor && a.activePanel == panelRecent {
			// Highlight full row
			writeLine(sb, lineRow, 1, BgCyan+FgBlack+strings.Repeat(" ", a.width)+Reset)
			writeLine(sb, lineRow, colName,    BgCyan+FgBlack+nameStr+Reset)
			writeLine(sb, lineRow, colCat,     BgCyan+FgBlack+catStr+Reset)
			writeLine(sb, lineRow, colTime,    BgCyan+FgBlack+timeStr+Reset)
			writeLine(sb, lineRow, colProfile, BgCyan+FgBlack+profileStr+Reset)
			writeLine(sb, lineRow, colCount,   BgCyan+FgBlack+countStr+Reset)
		} else {
			writeLine(sb, lineRow, colName,    Bold+nameStr+Reset)
			writeLine(sb, lineRow, colCat,     Dim+catStr+Reset)
			writeLine(sb, lineRow, colTime,    Dim+timeStr+Reset)
			writeLine(sb, lineRow, colProfile, Dim+profileStr+Reset)
			writeLine(sb, lineRow, colCount,   Dim+countStr+Reset)
		}
	}

	return startRow + 2 + maxRows + 1
}

func (a *App) renderFooter(sb *strings.Builder, row int) {
	writeLine(sb, row, 1, Dim+strings.Repeat("─", a.width)+Reset)
	row++

	if a.currentMode == modeSearch {
		writeLine(sb, row, 1, FgCyan+" Search: "+Reset+a.searchQuery+"█")
		if len(a.searchResults) > 0 {
			hint := fmt.Sprintf("  (%d results — Enter to open, Esc to cancel)", len(a.searchResults))
			writeLine(sb, row, 1+9+len(a.searchQuery)+1, Dim+hint+Reset)
		}
		row++
		for i, p := range a.searchResults {
			if i >= 5 {
				break
			}
			writeLine(sb, row, 3, FgCyan+Truncate(p.Name, 25)+Reset)
			writeLine(sb, row, 30, Dim+p.Category+Reset)
			row++
		}
	} else {
		help := Dim + " [↑↓] Navigate  [Tab] Switch panel  [Enter] Open VSCode  [t] Terminal tab  [/] Search  [p] Profile  [r] Recent  [q] Quit" + Reset
		writeLine(sb, row, 1, help)
		row++
		if a.statusMsg != "" {
			writeLine(sb, row, 2, FgGreen+a.statusMsg+Reset)
		}
	}
}

func (a *App) renderProfilePicker(sb *strings.Builder) {
	if a.pendingProject == nil {
		return
	}

	profiles := a.cfg.Profiles
	boxW := 46
	boxH := len(profiles) + 6

	startRow := (a.height-boxH)/2 + 1
	startCol := (a.width-boxW)/2 + 1
	if startRow < 1 {
		startRow = 1
	}
	if startCol < 1 {
		startCol = 1
	}

	inner := boxW - 2

	put := func(r int, s string) {
		sb.WriteString(MoveCursor(startRow+r, startCol) + s)
	}

	put(0, "╭"+strings.Repeat("─", inner)+"╮")
	put(1, "│ "+padTo(Bold+"Choose a profile"+Reset, inner-2)+" │")
	put(2, "│ "+padTo(Dim+Truncate(a.pendingProject.Name, inner-2)+Reset, inner-2)+" │")
	put(3, "├"+strings.Repeat("─", inner)+"┤")

	for i, p := range profiles {
		if i == a.profileCursor {
			put(4+i, "│ "+BgCyan+FgBlack+padTo(" > "+p.Name, inner-2)+Reset+" │")
		} else {
			put(4+i, "│ "+padTo(Dim+"   "+p.Name+Reset, inner-2)+" │")
		}
	}

	n := len(profiles)
	put(4+n, "├"+strings.Repeat("─", inner)+"┤")
	put(5+n, "│ "+Dim+padTo("[↑↓] Navigate  [Enter] Select  [Esc] Cancel", inner-2)+Reset+" │")
	put(6+n, "╰"+strings.Repeat("─", inner)+"╯")
}

func (a *App) panelTitle(title string, active bool) string {
	if active {
		return Bold + FgCyan + title + Reset
	}
	return Dim + title + Reset
}

// OpenDirect opens a project by name without TUI (CLI mode)
func OpenDirect(cfg *config.Config, hist *history.History, cats []project.Category, name string, profileOverride string) error {
	// Find project
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
		return fmt.Errorf("project %q not found", name)
	}

	profile := profileOverride
	if profile == "" {
		if saved, ok := hist.GetProfile(found.Path); ok {
			profile = saved
		} else {
			profile = cfg.GetProfileForCategory(found.Category)
		}
	}

	if !vscode.IsInstalled() {
		return fmt.Errorf("code binary not found in PATH")
	}

	if err := vscode.Open(found.Path, profile); err != nil {
		return err
	}

	hist.Record(found.Name, found.Path, found.Category, profile)
	history.Save(hist)

	fmt.Fprintf(os.Stdout, "✓ Opened %s with profile [%s]\n", found.Name, profile)
	return nil
}