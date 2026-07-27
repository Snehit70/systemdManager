from pathlib import Path


def replace_once(text: str, old: str, new: str, label: str) -> str:
    count = text.count(old)
    if count != 1:
        raise RuntimeError(f"{label}: expected one match, found {count}")
    return text.replace(old, new, 1)


def replace_count(text: str, old: str, new: str, expected: int, label: str) -> str:
    count = text.count(old)
    if count != expected:
        raise RuntimeError(f"{label}: expected {expected} matches, found {count}")
    return text.replace(old, new)


model_path = Path("internal/ui/model.go")
model = model_path.read_text()

model = replace_once(
    model,
    '''func (i item) Title() string       { return i.svc.Name }
func (i item) Description() string { return i.svc.Description }
func (i item) FilterValue() string { return i.svc.Name }

''',
    '''func (i item) Title() string       { return i.svc.Name }
func (i item) Description() string { return i.svc.Description }
func (i item) FilterValue() string { return i.svc.Name }

type listItemsRebuiltMsg struct {
\tmatches       list.FilterMatchesMsg
\tpreferred     string
\trefreshDetail bool
}

''',
    "list rebuild message type",
)

model = replace_once(
    model,
    '''\tcase list.FilterMatchesMsg:
\t\tm.list, cmd = m.list.Update(msg)
\t\treturn m, cmd
''',
    '''\tcase list.FilterMatchesMsg:
\t\tpreferred := m.selectedSvc
\t\tm.list, cmd = m.list.Update(msg)
\t\tselectionChanged := m.reconcileListSelection(preferred)
\t\tcmds = append(cmds, cmd)
\t\tif selectionChanged && m.selectedSvc != "" {
\t\t\tcmds = append(cmds, m.fetchDetailContent(m.selectedSvc))
\t\t}
\t\treturn m, tea.Batch(cmds...)

\tcase listItemsRebuiltMsg:
\t\tm.list, cmd = m.list.Update(msg.matches)
\t\tselectionChanged := m.reconcileListSelection(msg.preferred)
\t\tcmds = append(cmds, cmd)
\t\tif (selectionChanged || msg.refreshDetail) && m.selectedSvc != "" {
\t\t\tcmds = append(cmds, m.fetchDetailContent(m.selectedSvc))
\t\t}
\t\treturn m, tea.Batch(cmds...)
''',
    "filter result reconciliation",
)

model = replace_count(
    model,
    '\t\t\tcmds = append(cmds, m.list.SetItems(m.buildListItems()))',
    '\t\t\tcmds = append(cmds, m.replaceListItems(m.buildListItems(), false))',
    2,
    "group and source rebuild call sites",
)

model = replace_once(
    model,
    '''\t\titems := m.buildListItems()
\t\tcmds = append(cmds, m.list.SetItems(items))

\t\tif m.selectedSvc == "" && len(items) > 0 {
\t\t\tfor i, it := range items {
\t\t\t\tif _, ok := it.(item); ok {
\t\t\t\t\tm.list.Select(i)
\t\t\t\t\tm.selectedSvc = it.FilterValue()
\t\t\t\t\tbreak
\t\t\t\t}
\t\t\t}
\t\t}

\t\tif svc := m.getSelectedService(); svc != nil {
\t\t\tm.selectedSvc = svc.Name
\t\t\tcmds = append(cmds, m.fetchDetailContent(svc.Name))
\t\t}
''',
    '''\t\titems := m.buildListItems()
\t\tcmds = append(cmds, m.replaceListItems(items, true))
''',
    "service refresh rebuild",
)

model = replace_once(
    model,
    '''func (m MainModel) buildListItems() []list.Item {
''',
    '''func (m *MainModel) replaceListItems(items []list.Item, refreshDetail bool) tea.Cmd {
\tpreferred := m.selectedSvc
\tfilterCmd := m.list.SetItems(items)
\tif filterCmd != nil {
\t\treturn func() tea.Msg {
\t\t\tmsg := filterCmd()
\t\t\tmatches, ok := msg.(list.FilterMatchesMsg)
\t\t\tif !ok {
\t\t\t\treturn msg
\t\t\t}
\t\t\treturn listItemsRebuiltMsg{
\t\t\t\tmatches:       matches,
\t\t\t\tpreferred:     preferred,
\t\t\t\trefreshDetail: refreshDetail,
\t\t\t}
\t\t}
\t}

\tselectionChanged := m.reconcileListSelection(preferred)
\tif (selectionChanged || refreshDetail) && m.selectedSvc != "" {
\t\treturn m.fetchDetailContent(m.selectedSvc)
\t}
\treturn nil
}

func (m *MainModel) reconcileListSelection(preferred string) bool {
\tprevious := m.selectedSvc
\tselectedIndex := -1
\tselectedName := ""
\tvisibleItems := m.list.VisibleItems()

\tif preferred != "" {
\t\tfor i, listItem := range visibleItems {
\t\t\tserviceItem, ok := listItem.(item)
\t\t\tif ok && serviceItem.svc.Name == preferred {
\t\t\t\tselectedIndex = i
\t\t\t\tselectedName = serviceItem.svc.Name
\t\t\t\tbreak
\t\t\t}
\t\t}
\t}

\tif selectedIndex < 0 {
\t\tfor i, listItem := range visibleItems {
\t\t\tserviceItem, ok := listItem.(item)
\t\t\tif ok {
\t\t\t\tselectedIndex = i
\t\t\t\tselectedName = serviceItem.svc.Name
\t\t\t\tbreak
\t\t\t}
\t\t}
\t}

\tif selectedIndex >= 0 {
\t\tm.list.Select(selectedIndex)
\t\tm.selectedSvc = selectedName
\t} else {
\t\tm.list.ResetSelected()
\t\tm.selectedSvc = ""
\t\tm.viewport.SetContent("")
\t\tm.viewport.GotoTop()
\t}

\tif previous == m.selectedSvc {
\t\treturn false
\t}

\tif m.following {
\t\tm.stopFollow()
\t} else if m.followPending {
\t\tm.followSessionID++
\t\tm.followPending = false
\t}
\treturn true
}

func (m MainModel) buildListItems() []list.Item {
''',
    "selection helper insertion",
)

model_path.write_text(model)


test_path = Path("internal/ui/model_test.go")
tests = test_path.read_text()

tests = replace_once(
    tests,
    '''import (
\t"context"
\t"errors"
\t"os/exec"
\t"systemd-tui/internal/client"
\t"systemd-tui/internal/config"
\t"testing"

\ttea "github.com/charmbracelet/bubbletea"
)
''',
    '''import (
\t"context"
\t"errors"
\t"os/exec"
\t"strings"
\t"testing"

\t"systemd-tui/internal/client"
\t"systemd-tui/internal/config"

\ttea "github.com/charmbracelet/bubbletea"
)
''',
    "test imports",
)

tests = replace_once(
    tests,
    '''func newTestModel() MainModel {
\treturn NewMainModel(&mockServiceClient{}, config.Default())
}

''',
    '''func newTestModel() MainModel {
\treturn NewMainModel(&mockServiceClient{}, config.Default())
}

func applyServices(t *testing.T, model MainModel, services []client.Service) MainModel {
\tt.Helper()
\tupdated, _ := model.Update(services)
\treturn updated.(MainModel)
}

func selectedServiceName(t *testing.T, model MainModel) string {
\tt.Helper()
\tservice := model.getSelectedService()
\tif service == nil {
\t\tt.Fatal("expected a service to be selected")
\t}
\tif service.Name != model.selectedSvc {
\t\tt.Fatalf("visible selection %q does not match selectedSvc %q", service.Name, model.selectedSvc)
\t}
\treturn service.Name
}

''',
    "test helpers",
)

if "func TestGroupingPreservesSelectedService" in tests:
    raise RuntimeError("selection regression tests already exist")

tests += '''

func TestGroupingPreservesSelectedService(t *testing.T) {
\tmodel := newTestModel()
\tmodel = applyServices(t, model, []client.Service{
\t\t{Name: "active.service", Status: client.StatusActive, Source: client.SourceUser},
\t\t{Name: "inactive.service", Status: client.StatusInactive, Source: client.SourceUser},
\t\t{Name: "failed.service", Status: client.StatusFailed, Source: client.SourceUser},
\t})

\tmodel.list.Select(1)
\tmodel.selectedSvc = "inactive.service"
\tupdated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
\tmodel = updated.(MainModel)

\tif got := selectedServiceName(t, model); got != "inactive.service" {
\t\tt.Fatalf("selected service after grouping = %q, want inactive.service", got)
\t}
\tif _, ok := model.list.SelectedItem().(groupHeaderItem); ok {
\t\tt.Fatal("group header became the effective selection")
\t}
}

func TestSourceFilterFallsBackAndStopsFollow(t *testing.T) {
\tmodel := newTestModel()
\tmodel = applyServices(t, model, []client.Service{
\t\t{Name: "system.service", Status: client.StatusActive, Source: client.SourceSystem},
\t\t{Name: "user.service", Status: client.StatusActive, Source: client.SourceUser},
\t})

\tcancelled := false
\tmodel.following = true
\tmodel.followCancel = func() { cancelled = true }
\tupdated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("F")})
\tmodel = updated.(MainModel)

\tif got := selectedServiceName(t, model); got != "user.service" {
\t\tt.Fatalf("selected service after source filter = %q, want user.service", got)
\t}
\tif model.following {
\t\tt.Fatal("follow mode remained active after the selected service changed")
\t}
\tif !cancelled {
\t\tt.Fatal("active follow session was not cancelled")
\t}
}

func TestSourceFilterInvalidatesPendingFollow(t *testing.T) {
\tmodel := newTestModel()
\tmodel = applyServices(t, model, []client.Service{
\t\t{Name: "system.service", Status: client.StatusActive, Source: client.SourceSystem},
\t\t{Name: "user.service", Status: client.StatusActive, Source: client.SourceUser},
\t})

\tmodel.followPending = true
\tmodel.followSessionID = 7
\tupdated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("F")})
\tmodel = updated.(MainModel)

\tif model.followPending {
\t\tt.Fatal("pending follow remained active after the selected service changed")
\t}
\tif model.followSessionID != 8 {
\t\tt.Fatalf("follow session ID = %d, want 8", model.followSessionID)
\t}
}

func TestRefreshPreservesSelectionWhenOrderChanges(t *testing.T) {
\tmodel := newTestModel()
\tmodel = applyServices(t, model, []client.Service{
\t\t{Name: "alpha.service", Status: client.StatusActive, Source: client.SourceUser},
\t\t{Name: "beta.service", Status: client.StatusInactive, Source: client.SourceUser},
\t})

\tmodel.list.Select(1)
\tmodel.selectedSvc = "beta.service"
\tmodel = applyServices(t, model, []client.Service{
\t\t{Name: "beta.service", Status: client.StatusActive, Source: client.SourceUser},
\t\t{Name: "alpha.service", Status: client.StatusInactive, Source: client.SourceUser},
\t})

\tif got := selectedServiceName(t, model); got != "beta.service" {
\t\tt.Fatalf("selected service after refresh = %q, want beta.service", got)
\t}
\tif model.list.Index() != 0 {
\t\tt.Fatalf("selected index after reorder = %d, want 0", model.list.Index())
\t}
}

func TestEmptyRefreshClearsSelectionAndDetail(t *testing.T) {
\tmodel := newTestModel()
\tmodel = applyServices(t, model, []client.Service{
\t\t{Name: "demo.service", Status: client.StatusActive, Source: client.SourceUser},
\t})
\tmodel.viewport.Width = 40
\tmodel.viewport.Height = 5
\tmodel.viewport.SetContent("stale detail")
\tif !strings.Contains(model.viewport.View(), "stale detail") {
\t\tt.Fatal("test setup did not render stale detail content")
\t}

\tmodel = applyServices(t, model, nil)

\tif model.selectedSvc != "" {
\t\tt.Fatalf("selectedSvc after empty refresh = %q, want empty", model.selectedSvc)
\t}
\tif model.getSelectedService() != nil {
\t\tt.Fatal("a service remained selected after the list became empty")
\t}
\tif strings.Contains(model.viewport.View(), "stale detail") {
\t\tt.Fatal("stale detail content remained after the list became empty")
\t}
}

func TestInitialGroupedListSkipsHeader(t *testing.T) {
\tmodel := newTestModel()
\tmodel.groupMode = groupByStatus
\tmodel = applyServices(t, model, []client.Service{
\t\t{Name: "demo.service", Status: client.StatusActive, Source: client.SourceUser},
\t})

\tif got := selectedServiceName(t, model); got != "demo.service" {
\t\tt.Fatalf("selected service = %q, want demo.service", got)
\t}
\tif model.list.Index() != 1 {
\t\tt.Fatalf("selected index = %d, want 1 after the group header", model.list.Index())
\t}
}
'''

test_path.write_text(tests)
