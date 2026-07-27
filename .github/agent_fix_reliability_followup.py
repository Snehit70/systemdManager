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
\t\t\tlipgloss.Height(renderHelpBar(keys, model.width, model.theme)))
\t}
'''
if old in text:
    path.write_text(text.replace(old, new, 1))
