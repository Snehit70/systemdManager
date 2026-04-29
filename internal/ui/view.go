package ui

import (
	"fmt"
	"strings"

	"systemd-tui/internal/config"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// modalLabelWidth is the fixed left-column width for create modal labels.
const modalLabelWidth = 13

func (m MainModel) View() string {
	var listStyle, detailStyle lipgloss.Style

	if m.activeView == listView {
		listStyle = m.activeBorder
		detailStyle = m.inactiveBorder
	} else {
		listStyle = m.inactiveBorder
		detailStyle = m.activeBorder
	}

	listView := listStyle.Render(m.list.View())
	detailView := detailStyle.Render(m.viewport.View())

	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listView,
		detailView,
	)

	fullWidth := m.width
	if fullWidth < 1 {
		fullWidth = lipgloss.Width(mainView)
	}

	statusBar := m.renderStatusBar(fullWidth)

	filterBar := ""
	if m.list.FilterState() == list.Filtering {
		filterBar = m.renderFilterBar(fullWidth)
	}

	var helpView string
	if m.help.ShowAll {
		helpView = lipgloss.NewStyle().
			Width(fullWidth).
			PaddingLeft(1).
			Background(lipgloss.Color(m.theme.SurfaceDeep)).
			Render(m.help.View(keys))
	} else {
		helpView = renderHelpBar(keys, fullWidth, m.theme)
	}

	baseView := lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		filterBar,
		statusBar,
		helpView,
	)

	if m.showCreate {
		return m.renderCreateModal(baseView)
	}

	return baseView
}

// --- Create modal ----------------------------------------------------------

func (m MainModel) renderModalLabel(label string, index int) string {
	style := lipgloss.NewStyle().Width(modalLabelWidth).Bold(true)
	if m.createModal.focusIndex == index {
		return style.Foreground(lipgloss.Color(m.theme.BorderActive)).Render(label)
	}
	return style.Foreground(lipgloss.Color(m.theme.TextMuted)).Render(label)
}

func (m MainModel) renderCreateModal(baseView string) string {
	modalWidth := 64
	if m.width > 0 && modalWidth > m.width-8 {
		modalWidth = m.width - 8
	}
	if modalWidth < 54 {
		modalWidth = 54
	}
	innerWidth := modalWidth - 4 // padding(1,2) -> 2 left + 2 right
	fieldWidth := innerWidth - modalLabelWidth - 1
	if fieldWidth < 16 {
		fieldWidth = 16
	}

	subtle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.TextMuted))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Accent)).Bold(true)
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.BorderActive)).
		Bold(true).
		Render("✦  Create New Service")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, subtle.Render("Creates a user unit in ~/.config/systemd/user/"))
	lines = append(lines, "")
	lines = append(lines, m.renderModalTextInput("Name", 0, &m.createModal.nameInput, fieldWidth))
	lines = append(lines, m.renderModalTextInput("Command", 1, &m.createModal.execInput, fieldWidth))
	lines = append(lines, m.renderModalTextInput("Description", 2, &m.createModal.descInput, fieldWidth))
	lines = append(lines, m.renderModalTextInput("Workdir", 3, &m.createModal.workdirInput, fieldWidth))
	lines = append(lines, m.renderModalChoiceField("Type", 4, m.createModal.serviceType, "t", fieldWidth))
	lines = append(lines, m.renderModalChoiceField("Restart", 5, m.createModal.restart, "r", fieldWidth))
	lines = append(lines, "")

	if m.creating {
		lines = append(lines, accent.Render("⠋  Creating service…"))
	} else {
		lines = append(lines, joinHints([]hint{
			{key: "tab", desc: "next"},
			{key: "↵", desc: "create"},
			{key: "esc", desc: "cancel"},
		}, m.theme))
	}

	// Pre-pad every line to the inner width so the modal background paints
	// uniformly without leaving the trailing dark bands seen in the old UI.
	filled := make([]string, len(lines))
	for i, l := range lines {
		filled[i] = padBgLine(l, innerWidth, m.theme.SurfaceDeep)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, filled...)

	content := lipgloss.NewStyle().
		Padding(1, 2).
		Background(lipgloss.Color(m.theme.SurfaceDeep)).
		Render(body)

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.BorderActive)).
		BorderBackground(lipgloss.Color(m.theme.SurfaceDeep)).
		Render(content)

	return overlayCenter(baseView, modal, m.width, m.height)
}

func (m MainModel) renderModalTextInput(label string, index int, input *textinput.Model, fieldWidth int) string {
	focused := m.createModal.focusIndex == index
	bgColor := m.theme.Surface
	borderColor := m.theme.Border
	if focused {
		borderColor = m.theme.BorderActive
	}

	input.Width = fieldWidth - 2
	input.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Text)).
		Background(lipgloss.Color(bgColor))
	input.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.TextDim)).
		Background(lipgloss.Color(bgColor))
	input.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.SurfaceDeep)).
		Background(lipgloss.Color(m.theme.BorderActive))

	field := lipgloss.NewStyle().
		Background(lipgloss.Color(bgColor)).
		Foreground(lipgloss.Color(m.theme.Text)).
		Padding(0, 1).
		Width(fieldWidth).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(borderColor)).
		BorderBackground(lipgloss.Color(m.theme.SurfaceDeep)).
		Render(input.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, m.renderModalLabel(label, index), " ", field)
}

func (m MainModel) renderModalChoiceField(label string, index int, value, key string, fieldWidth int) string {
	focused := m.createModal.focusIndex == index
	valColor := m.theme.Text
	if focused {
		valColor = m.theme.BorderActive
	}
	val := lipgloss.NewStyle().
		Foreground(lipgloss.Color(valColor)).
		Bold(true).
		Render(value)

	hintText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.TextDim)).
		Render(" cycle")
	rest := val + "   " + kbdChip(key, m.theme) + hintText

	field := lipgloss.NewStyle().Padding(0, 1).Width(fieldWidth).Render(rest)
	return lipgloss.JoinHorizontal(lipgloss.Top, m.renderModalLabel(label, index), " ", field)
}

// --- Detail pane header ---------------------------------------------------

// renderDetailHeader renders the service identity block for the detail pane:
// the service name in bold, status + source pills, the description, and a
// banner separator labelling the section (e.g. "Last 50 lines" or
// "FOLLOWING").
//
// `width` is the viewport content width; passing 0 disables the trailing rule.
func renderDetailHeader(theme config.ThemeColors, name, status, sub, source, description, banner string, width int) string {
	statusLabel := status
	if sub != "" && sub != status {
		statusLabel = status + " " + sub
	}

	statusPill := statusBadge(theme, status, statusLabel)
	sourcePill := sourceBadge(theme, source)

	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Text)).
		Bold(true)

	pills := statusPill + "  " + sourcePill
	pillsWidth := lipgloss.Width(pills)
	nameRendered := nameStyle.Render(name)
	nameWidth := lipgloss.Width(nameRendered)

	var headerLine string
	if width > 0 && nameWidth+pillsWidth+2 < width {
		gap := width - nameWidth - pillsWidth
		if gap < 1 {
			gap = 1
		}
		headerLine = nameRendered + strings.Repeat(" ", gap) + pills
	} else {
		headerLine = nameRendered + "  " + pills
	}

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextMuted))
	descLine := descStyle.Render(description)

	bannerLine := renderSectionBanner(theme, banner, width)

	return strings.Join([]string{headerLine, descLine, "", bannerLine, ""}, "\n")
}

func statusBadge(theme config.ThemeColors, status, label string) string {
	color := theme.StatusInactive
	switch status {
	case "active":
		color = theme.StatusActive
	case "failed":
		color = theme.StatusFailed
	case "activating", "deactivating", "reloading":
		color = theme.StatusActivating
	}
	dot := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("●")
	text := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Render(label)
	return dot + " " + text
}

func sourceBadge(theme config.ThemeColors, source string) string {
	color := theme.SourceStatic
	switch source {
	case "user":
		color = theme.SourceUser
	case "system":
		color = theme.SourceSystem
	case "transient":
		color = theme.SourceTransient
	case "generated":
		color = theme.SourceGenerated
	case "static":
		color = theme.SourceStatic
	}
	if source == "" {
		source = "unknown"
	}
	dot := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("◆")
	text := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextMuted)).
		Render(source)
	return dot + " " + text
}

func renderSectionBanner(theme config.ThemeColors, label string, width int) string {
	if label == "" {
		return ""
	}
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Accent)).
		Bold(true)
	rule := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Border))

	rendered := labelStyle.Render(label)
	if width <= 0 {
		return rule.Render("──── ") + rendered
	}
	prefix := rule.Render("──── ")
	used := lipgloss.Width(prefix) + lipgloss.Width(rendered) + 1
	if used >= width {
		return prefix + rendered
	}
	return prefix + rendered + " " + rule.Render(strings.Repeat("─", width-used))
}

// followBanner renders the FOLLOWING section banner with a pulsing-style
// indicator. Bubble Tea is one-shot per render, so the indicator is static
// but visually distinct.
func followBanner(theme config.ThemeColors, width int) string {
	indicator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.StatusActive)).
		Bold(true).
		Render("⦿ FOLLOWING")
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextMuted)).
		Render("press f to stop")
	rule := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Border))

	left := indicator + "  " + hint
	if width <= 0 {
		return left
	}
	used := lipgloss.Width(left) + 1
	if used >= width {
		return left
	}
	return left + " " + rule.Render(strings.Repeat("─", width-used))
}

// --- Overlay helper --------------------------------------------------------

func overlayCenter(base, modal string, width, height int) string {
	if width <= 0 || height <= 0 {
		return modal
	}

	baseLines := strings.Split(base, "\n")
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	modalLines := strings.Split(modal, "\n")
	modalWidth := 0
	for _, line := range modalLines {
		if lineWidth := ansi.StringWidth(line); lineWidth > modalWidth {
			modalWidth = lineWidth
		}
	}
	modalHeight := len(modalLines)
	x := (width - modalWidth) / 2
	y := (height - modalHeight) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	for i, modalLine := range modalLines {
		target := y + i
		if target < 0 || target >= len(baseLines) {
			continue
		}

		line := baseLines[target]
		left := ansi.Cut(line, 0, x)
		right := ansi.Cut(line, x+modalWidth, width)
		baseLines[target] = left + modalLine + right
	}

	if len(baseLines) > height {
		baseLines = baseLines[:height]
	}
	return strings.Join(baseLines, "\n")
}

// --- Detail viewport scroll info ------------------------------------------

func (m *MainModel) viewportWithScrollInfo(content string) string {
	totalLines := strings.Count(content, "\n") + 1
	viewportHeight := m.viewport.Height
	showScroll := totalLines > viewportHeight && viewportHeight > 0

	if !showScroll {
		return content
	}

	scrollPos := m.viewport.YOffset
	return fmt.Sprintf("[%d/%d]\n%s", scrollPos, totalLines, content)
}

// --- Status, filter and help bars -----------------------------------------

func (m MainModel) renderStatusBar(width int) string {
	bg := m.theme.Surface

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Text)).
		Background(lipgloss.Color(bg)).
		Bold(true)
	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.TextMuted)).
		Background(lipgloss.Color(bg))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Border)).
		Background(lipgloss.Color(bg))
	sep := sepStyle.Render("  ")

	var segments []string

	filtered := m.filteredServices()
	totalCount := len(m.services)
	if len(filtered) != totalCount {
		segments = append(segments, countStyle.Render(fmt.Sprintf("%d", len(filtered)))+
			dimStyle.Render(fmt.Sprintf("/%d services", totalCount)))
	} else {
		segments = append(segments, countStyle.Render(fmt.Sprintf("%d", totalCount))+
			dimStyle.Render(" services"))
	}

	if m.filterMode != filterAll {
		segments = append(segments, chip("filter", m.filterMode.String(), m.theme.Accent, m.theme))
	}
	if m.groupMode != groupNone {
		segments = append(segments, chip("group", m.groupMode.String(), m.theme.SourceGenerated, m.theme))
	}
	if m.detailViewMode != detailViewLogs {
		segments = append(segments, chip("view", m.detailViewMode.String(), m.theme.SourceTransient, m.theme))
	}
	if m.following {
		segments = append(segments, chip("●", "following", m.theme.StatusActive, m.theme))
	}

	left := strings.Join(segments, sep)

	right := ""
	if m.statusMessage != "" {
		right = dimStyle.Render(m.statusMessage)
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := 2

	bar := left
	if width > 0 && leftWidth+rightWidth+padding < width {
		gap := width - leftWidth - rightWidth - padding
		if gap < 1 {
			gap = 1
		}
		bar = left + lipgloss.NewStyle().Background(lipgloss.Color(bg)).Render(strings.Repeat(" ", gap)) + right
	}

	return lipgloss.NewStyle().
		PaddingLeft(1).
		Width(width).
		Background(lipgloss.Color(bg)).
		Render(bar)
}

func (m MainModel) renderFilterBar(width int) string {
	bg := m.theme.Surface
	prompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.BorderActive)).
		Background(lipgloss.Color(bg)).
		Bold(true).
		Render(" / ")

	// Style the underlying filter input to match the bar background.
	m.list.FilterInput.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Text)).
		Background(lipgloss.Color(bg))
	m.list.FilterInput.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.TextDim)).
		Background(lipgloss.Color(bg))
	m.list.FilterInput.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.SurfaceDeep)).
		Background(lipgloss.Color(m.theme.BorderActive))

	body := prompt + m.list.FilterInput.View()

	return lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color(bg)).
		Render(body)
}

func renderHelpBar(k keyMap, width int, theme config.ThemeColors) string {
	bg := theme.SurfaceDeep
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextMuted)).
		Background(lipgloss.Color(bg))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Border)).
		Background(lipgloss.Color(bg))

	row1 := []key.Binding{k.Up, k.Down, k.Filter, k.SwitchFocus, k.ToggleFollow, k.ToggleDetail}
	row2 := []key.Binding{k.Start, k.Stop, k.Restart, k.Edit, k.Enable, k.Disable, k.ToggleGroup, k.ToggleSource, k.Create, k.Help, k.Quit}

	sep := sepStyle.Render(" · ")

	formatRow := func(bs []key.Binding) string {
		parts := make([]string, 0, len(bs))
		for _, b := range bs {
			h := b.Help()
			parts = append(parts, kbdChipBg(h.Key, theme, bg)+" "+descStyle.Render(h.Desc))
		}
		return strings.Join(parts, sep)
	}

	line1 := formatRow(row1)
	line2 := formatRow(row2)

	rowStyle := lipgloss.NewStyle().
		Width(width).
		PaddingLeft(1).
		Background(lipgloss.Color(bg))

	return lipgloss.JoinVertical(lipgloss.Left,
		rowStyle.Render(line1),
		rowStyle.Render(line2),
	)
}

// --- Small style helpers --------------------------------------------------

type hint struct {
	key  string
	desc string
}

func joinHints(hints []hint, theme config.ThemeColors) string {
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextMuted))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Border))
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, kbdChip(h.key, theme)+" "+descStyle.Render(h.desc))
	}
	return strings.Join(parts, sepStyle.Render("  ·  "))
}

// kbdChip renders a key as a small key-cap style chip on a transparent
// background.
func kbdChip(label string, theme config.ThemeColors) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Text)).
		Background(lipgloss.Color(theme.Surface)).
		Bold(true).
		Padding(0, 1).
		Render(label)
}

// kbdChipBg renders a key chip with an explicit outer background so the chip
// sits inside a filled bar without ghosting.
func kbdChipBg(label string, theme config.ThemeColors, _ string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Accent)).
		Background(lipgloss.Color(theme.Surface)).
		Bold(true).
		Padding(0, 1).
		Render(label)
}

// chip renders a "label value" pill, e.g. `filter all` or `view status`.
func chip(label, value, accent string, theme config.ThemeColors) string {
	bg := theme.Surface
	labelPart := lipgloss.NewStyle().
		Foreground(lipgloss.Color(accent)).
		Background(lipgloss.Color(bg)).
		Bold(true).
		Render(label)
	valPart := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Text)).
		Background(lipgloss.Color(bg)).
		Render(value)
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Padding(0, 1).
		Render(labelPart + " " + valPart)
}

// padBgLine right-pads `s` to `width` columns with a uniform background. It
// preserves any styled content already in `s` and only paints the trailing
// gap, which avoids the dark banding caused by nested bg styles.
func padBgLine(s string, width int, bg string) string {
	curr := ansi.StringWidth(s)
	if curr >= width {
		return s
	}
	pad := lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Render(strings.Repeat(" ", width-curr))
	return s + pad
}

// padToWidth right-pads a string to width without applying any background.
func padToWidth(s string, width int) string {
	curr := ansi.StringWidth(s)
	if curr >= width {
		return s
	}
	return s + strings.Repeat(" ", width-curr)
}
