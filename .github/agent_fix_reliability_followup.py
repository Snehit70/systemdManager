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
