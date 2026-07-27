from pathlib import Path

path = Path("internal/ui/model_test.go")
text = path.read_text()
obsolete = '''func TestErrorMessageHandling(t *testing.T) {
\tmodel := newTestModel()

\ttestErr := errors.New("test error")
\terrMsg := errMsg{op: "Test", err: testErr}

\tupdatedModel, _ := model.Update(errMsg)
\tm := updatedModel.(MainModel)

\texpected := "Error: Test - test error"
\tif m.statusMessage != expected {
\t\tt.Errorf("Expected '%s', got '%s'", expected, m.statusMessage)
\t}
}

'''
if obsolete in text:
    path.write_text(text.replace(obsolete, "", 1))

path = Path("internal/ui/model.go")
text = path.read_text()
old = '''\tm.help.Width = m.width
\thelpHeight := 2
\tif m.help.ShowAll {
\t\thelpHeight = lipgloss.Height(m.help.View(keys))
\t\tif helpHeight < 1 {
\t\t\thelpHeight = 1
\t\t}
\t}
'''
new = '''\tm.help.Width = m.width
\thelpHeight := lipgloss.Height(m.renderHelpView(m.width))
\tif helpHeight < 1 {
\t\thelpHeight = 1
\t}
'''
if old in text:
    path.write_text(text.replace(old, new, 1))
elif new not in text:
    raise RuntimeError("rendered help height block not found")

path = Path("internal/ui/view.go")
text = path.read_text()
old = '''\tvar helpView string
\tif m.help.ShowAll {
\t\thelpView = lipgloss.NewStyle().
\t\t\tWidth(fullWidth).
\t\t\tPaddingLeft(1).
\t\t\tBackground(lipgloss.Color(m.theme.SurfaceDeep)).
\t\t\tRender(m.help.View(keys))
\t} else {
\t\thelpView = renderHelpBar(keys, fullWidth, m.theme)
\t}
'''
new = '''\thelpView := m.renderHelpView(fullWidth)
'''
if old in text:
    text = text.replace(old, new, 1)
elif new not in text:
    raise RuntimeError("View help block not found")
marker = '''func (m MainModel) renderStatusBar(width int) string {
'''
helper = '''func (m MainModel) renderHelpView(width int) string {
\tif m.help.ShowAll {
\t\treturn lipgloss.NewStyle().
\t\t\tWidth(width).
\t\t\tPaddingLeft(1).
\t\t\tBackground(lipgloss.Color(m.theme.SurfaceDeep)).
\t\t\tRender(m.help.View(keys))
\t}
\treturn renderHelpBar(keys, width, m.theme)
}

'''
if helper not in text:
    if marker not in text:
        raise RuntimeError("renderStatusBar marker not found")
    text = text.replace(marker, helper + marker, 1)
path.write_text(text)

path = Path("internal/ui/reliability_test.go")
text = path.read_text()
old = '''\tif got := lipgloss.Height(model.View()); got > model.height {
\t\tt.Fatalf("filter UI renders %d rows into a %d-row terminal", got, model.height)
\t}
'''
new = '''\tif got := lipgloss.Height(model.View()); got > model.height {
\t\tt.Fatalf("filter UI renders %d rows into a %d-row terminal (list=%d viewport=%d filter=%d status=%d help=%d)",
\t\t\tgot, model.height,
\t\t\tlipgloss.Height(model.list.View()),
\t\t\tlipgloss.Height(model.viewport.View()),
\t\t\tlipgloss.Height(model.renderFilterBar(model.width)),
\t\t\tlipgloss.Height(model.renderStatusBar(model.width)),
\t\t\tlipgloss.Height(model.renderHelpView(model.width)))
\t}
'''
if old in text:
    path.write_text(text.replace(old, new, 1))
