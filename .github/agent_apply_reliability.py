from pathlib import Path


def replace_once(text: str, old: str, new: str, label: str) -> str:
    count = text.count(old)
    if count != 1:
        raise RuntimeError(f"{label}: expected one match, found {count}")
    return text.replace(old, new, 1)


def update(path: str, transform):
    file_path = Path(path)
    original = file_path.read_text()
    updated = transform(original)
    if updated != original:
        file_path.write_text(updated)


def update_client_types(text: str) -> str:
    text = replace_once(
        text,
        '''type LogOptions struct {
\tLines  int    // Number of lines to fetch (default: 50)
\tFollow bool   // Whether to follow/stay attached (live tailing)
\tFilter string // Optional log level or text filter
}
''',
        '''type LogOptions struct {
\tLines  int    // Number of lines to fetch (default: 50)
\tFollow bool   // Whether to follow/stay attached (live tailing)
\tFilter string // Optional log level or text filter
}

// LogEvent carries either one live log line or a terminal stream error.
type LogEvent struct {
\tLine string
\tErr  error
}
''',
        "add LogEvent",
    )
    text = replace_once(
        text,
        '''\t// FollowLogs returns a channel that emits log lines.
\t// Call the returned cancel function to stop the stream.
\tFollowLogs(ctx context.Context, name string, opts LogOptions) (<-chan string, context.CancelFunc, error)
''',
        '''\t// FollowLogs returns a channel that emits log lines and terminal stream errors.
\t// Call the returned cancel function to stop the stream; cancellation is not emitted as an error.
\tFollowLogs(ctx context.Context, name string, opts LogOptions) (<-chan LogEvent, context.CancelFunc, error)
''',
        "update FollowLogs interface",
    )
    return text


def update_service_client(text: str) -> str:
    text = replace_once(text, 'import (\n\t"bufio"\n', 'import (\n\t"bufio"\n\t"bytes"\n', "add bytes import")
    text = replace_once(text, '\t"regexp"\n\t"strings"\n', '\t"regexp"\n\t"sort"\n\t"strings"\n', "add sort import")
    text = replace_once(
        text,
        '''func (c *systemdClient) ListServices(ctx context.Context) ([]client.Service, error) {
\tunits, err := c.listUnits(ctx)
\tif err != nil {
\t\treturn nil, err
\t}

\tunitFileStates := make(map[string]string)
\tunitFiles, err := c.listUnitFiles(ctx)
\tif err == nil {
\t\tfor _, uf := range unitFiles {
\t\t\tunitFileStates[uf.UnitFile] = uf.State
\t\t}
\t}

\tservices := make([]client.Service, len(units))
\tfor i, u := range units {
\t\tservices[i] = c.unitToService(u, unitFileStates[u.Unit])
\t}

\treturn services, nil
}
''',
        '''func (c *systemdClient) ListServices(ctx context.Context) ([]client.Service, error) {
\tunits, err := c.listUnits(ctx)
\tif err != nil {
\t\treturn nil, err
\t}

\tunitFiles, err := c.listUnitFiles(ctx)
\tif err != nil {
\t\treturn nil, err
\t}

\treturn c.mergeServices(units, unitFiles), nil
}
''',
        "replace service enumeration",
    )
    marker = '''func (c *systemdClient) StartService(ctx context.Context, name string) error {
'''
    helper = '''func (c *systemdClient) mergeServices(units []Unit, unitFiles []UnitFile) []client.Service {
\tunitFileStates := make(map[string]string, len(unitFiles))
\tfor _, unitFile := range unitFiles {
\t\tunitFileStates[unitFile.UnitFile] = unitFile.State
\t}

\tservicesByName := make(map[string]client.Service, len(units)+len(unitFiles))
\tfor _, unit := range units {
\t\tservicesByName[unit.Unit] = c.unitToService(unit, unitFileStates[unit.Unit])
\t}

\tfor _, unitFile := range unitFiles {
\t\tif !strings.HasSuffix(unitFile.UnitFile, ".service") || isTemplateUnitFile(unitFile.UnitFile) {
\t\t\tcontinue
\t\t}
\t\tif _, loaded := servicesByName[unitFile.UnitFile]; loaded {
\t\t\tcontinue
\t\t}

\t\tservicesByName[unitFile.UnitFile] = client.Service{
\t\t\tName:    unitFile.UnitFile,
\t\t\tStatus:  client.StatusInactive,
\t\t\tSub:     "dead",
\t\t\tEnabled: isUnitFileEnabled(unitFile.State),
\t\t\tLoad:    "unloaded",
\t\t\tSource:  c.determineSource(unitFile.UnitFile, unitFile.State),
\t\t}
\t}

\tnames := make([]string, 0, len(servicesByName))
\tfor name := range servicesByName {
\t\tnames = append(names, name)
\t}
\tsort.Strings(names)

\tservices := make([]client.Service, 0, len(names))
\tfor _, name := range names {
\t\tservices = append(services, servicesByName[name])
\t}
\treturn services
}

func isTemplateUnitFile(name string) bool {
\treturn strings.HasSuffix(name, "@.service")
}

'''
    text = replace_once(text, marker, helper + marker, "insert service merge helper")
    text = replace_once(
        text,
        '''func (c *systemdClient) FollowLogs(ctx context.Context, name string, opts client.LogOptions) (<-chan string, context.CancelFunc, error) {
\tlines := opts.Lines
\tif lines <= 0 {
\t\tlines = 50
\t}

\tctx, cancel := context.WithCancel(ctx)
\tcmd := exec.CommandContext(ctx, "journalctl", "--user", "-u", name, "-n", fmt.Sprintf("%d", lines), "-f", "--no-pager")

\tstdout, err := cmd.StdoutPipe()
\tif err != nil {
\t\tcancel()
\t\treturn nil, nil, fmt.Errorf("failed to create pipe for %s: %w", name, err)
\t}

\tif err := cmd.Start(); err != nil {
\t\tcancel()
\t\treturn nil, nil, fmt.Errorf("failed to start journalctl for %s: %w", name, err)
\t}

\tlogChan := make(chan string, 100)

\tgo func() {
\t\tdefer close(logChan)
\t\tdefer cmd.Wait()

\t\tscanner := bufio.NewScanner(stdout)
\t\tscanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
\t\tfor scanner.Scan() {
\t\t\tselect {
\t\t\tcase logChan <- scanner.Text():
\t\t\tcase <-ctx.Done():
\t\t\t\treturn
\t\t\t}
\t\t}
\t}()

\treturn logChan, cancel, nil
}
''',
        '''func (c *systemdClient) FollowLogs(ctx context.Context, name string, opts client.LogOptions) (<-chan client.LogEvent, context.CancelFunc, error) {
\tlines := opts.Lines
\tif lines <= 0 {
\t\tlines = 50
\t}

\tctx, cancel := context.WithCancel(ctx)
\tcmd := exec.CommandContext(ctx, "journalctl", "--user", "-u", name, "-n", fmt.Sprintf("%d", lines), "-f", "--no-pager")

\tstdout, err := cmd.StdoutPipe()
\tif err != nil {
\t\tcancel()
\t\treturn nil, nil, fmt.Errorf("failed to create pipe for %s: %w", name, err)
\t}
\tvar stderr bytes.Buffer
\tcmd.Stderr = &stderr

\tif err := cmd.Start(); err != nil {
\t\tcancel()
\t\treturn nil, nil, fmt.Errorf("failed to start journalctl for %s: %w", name, err)
\t}

\tlogChan := make(chan client.LogEvent, 100)

\tgo func() {
\t\tdefer close(logChan)

\t\tsend := func(event client.LogEvent) bool {
\t\t\tselect {
\t\t\tcase logChan <- event:
\t\t\t\treturn true
\t\t\tcase <-ctx.Done():
\t\t\t\treturn false
\t\t\t}
\t\t}

\t\tscanner := bufio.NewScanner(stdout)
\t\tscanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
\t\tfor scanner.Scan() {
\t\t\tif !send(client.LogEvent{Line: scanner.Text()}) {
\t\t\t\treturn
\t\t\t}
\t\t}

\t\tscanErr := scanner.Err()
\t\tif scanErr != nil && cmd.Process != nil {
\t\t\t_ = cmd.Process.Kill()
\t\t}
\t\twaitErr := cmd.Wait()
\t\tif ctx.Err() != nil {
\t\t\treturn
\t\t}

\t\tvar streamErrs []error
\t\tif scanErr != nil {
\t\t\tstreamErrs = append(streamErrs, fmt.Errorf("failed to read journalctl output for %s: %w", name, scanErr))
\t\t}
\t\tif waitErr != nil {
\t\t\tstreamErrs = append(streamErrs, commandError(fmt.Sprintf("journalctl follow failed for %s", name), waitErr, stderr.Bytes()))
\t\t}
\t\tif streamErr := errors.Join(streamErrs...); streamErr != nil {
\t\t\tsend(client.LogEvent{Err: streamErr})
\t\t}
\t}()

\treturn logChan, cancel, nil
}
''',
        "replace FollowLogs implementation",
    )
    text = replace_once(
        text,
        '''\tswitch state {
\tcase "generated":
\t\treturn client.SourceGenerated
\tcase "transient":
\t\treturn client.SourceTransient
\tcase "static", "alias":
\t\treturn client.SourceStatic
\tcase "enabled", "disabled":
\t\treturn client.SourceSystem
\tdefault:
\t\treturn client.SourceUnknown
\t}
''',
        '''\tswitch state {
\tcase "generated":
\t\treturn client.SourceGenerated
\tcase "transient":
\t\treturn client.SourceTransient
\tcase "static", "alias", "indirect":
\t\treturn client.SourceStatic
\tcase "enabled", "enabled-runtime", "disabled", "disabled-runtime", "linked", "linked-runtime", "masked", "masked-runtime":
\t\treturn client.SourceSystem
\tdefault:
\t\treturn client.SourceUnknown
\t}
''',
        "expand source states",
    )
    return text


def update_commands(text: str) -> str:
    text = replace_once(
        text,
        '''type errMsg struct {
\top  string
\terr error
}

func (e errMsg) Error() string { return e.err.Error() }

''',
        '''type serviceListMsg struct {
\tid       uint64
\tservices []client.Service
\terr      error
}

''',
        "replace generic error message",
    )
    text = replace_once(text, '\tch     <-chan string\n', '\tch     <-chan client.LogEvent\n', "update follow channel type")
    text = replace_once(
        text,
        '''type followStoppedMsg struct {
\tid   uint64
\tunit string
}
''',
        '''type followStoppedMsg struct {
\tid   uint64
\tunit string
}

type followStartFailedMsg struct {
\tid   uint64
\tunit string
\terr  error
}

type followStreamFailedMsg struct {
\tid   uint64
\tunit string
\terr  error
}
''',
        "add scoped follow failures",
    )
    text = replace_once(
        text,
        '''func (m MainModel) fetchServices() tea.Msg {
\tservices, err := m.client.ListServices(m.ctx)
\tif err != nil {
\t\treturn errMsg{op: "Failed to list services", err: err}
\t}
\treturn services
}
''',
        '''func (m MainModel) fetchServices(id uint64) tea.Cmd {
\treturn func() tea.Msg {
\t\tservices, err := m.client.ListServices(m.ctx)
\t\treturn serviceListMsg{id: id, services: services, err: err}
\t}
}
''',
        "scope service refresh",
    )
    text = replace_once(
        text,
        '''\t\tlogChan, cancel, err := m.client.FollowLogs(m.ctx, unit, client.LogOptions{Lines: lines})
\t\tif err != nil {
\t\t\treturn errMsg{op: "Failed to follow logs", err: err}
\t\t}
''',
        '''\t\tlogChan, cancel, err := m.client.FollowLogs(m.ctx, unit, client.LogOptions{Lines: lines})
\t\tif err != nil {
\t\t\treturn followStartFailedMsg{id: id, unit: unit, err: err}
\t\t}
''',
        "scope follow start error",
    )
    text = replace_once(
        text,
        '''\treturn func() tea.Msg {
\t\tselect {
\t\tcase line, ok := <-ch:
\t\t\tif !ok {
\t\t\t\treturn followStoppedMsg{id: id, unit: unit}
\t\t\t}
\t\t\treturn logLineMsg{id: id, line: line}
\t\tcase <-ctx.Done():
\t\t\treturn followStoppedMsg{id: id, unit: unit}
\t\t}
\t}
''',
        '''\treturn func() tea.Msg {
\t\tselect {
\t\tcase event, ok := <-ch:
\t\t\tif !ok {
\t\t\t\treturn followStoppedMsg{id: id, unit: unit}
\t\t\t}
\t\t\tif event.Err != nil {
\t\t\t\treturn followStreamFailedMsg{id: id, unit: unit, err: event.Err}
\t\t\t}
\t\t\treturn logLineMsg{id: id, line: event.Line}
\t\tcase <-ctx.Done():
\t\t\treturn followStoppedMsg{id: id, unit: unit}
\t\t}
\t}
''',
        "handle follow events",
    )
    return text


def update_model(text: str) -> str:
    text = replace_once(text, '\tfollowLogChan   <-chan string\n', '\tfollowLogChan   <-chan client.LogEvent\n', "update model follow channel")
    text = replace_once(
        text,
        '''\tactiveView    int
\tctx           context.Context

\tconfirmingAction string
''',
        '''\tactiveView    int
\tctx           context.Context

\trefreshRequestID uint64
\trefreshPending   bool

\tconfirmingAction string
''',
        "add refresh state",
    )
    text = replace_once(
        text,
        '''\t\tctx:            context.Background(),
\t\tsplitRatio:     0.33,
''',
        '''\t\tctx:              context.Background(),
\t\trefreshRequestID: 1,
\t\trefreshPending:   true,
\t\tsplitRatio:       0.33,
''',
        "initialize refresh state",
    )
    text = replace_once(
        text,
        '''func (m MainModel) Init() tea.Cmd {
\treturn tea.Batch(m.fetchServices, m.tick())
}
''',
        '''func (m MainModel) Init() tea.Cmd {
\treturn tea.Batch(m.fetchServices(m.refreshRequestID), m.tick())
}
''',
        "update initial refresh",
    )
    text = replace_once(
        text,
        '''func (m *MainModel) recalcPanelSizes() {
\thelpHeight := 2
\tstatusBarHeight := 1
\tmainHeight := m.height - helpHeight - statusBarHeight

\tlistWidth := int(float64(m.width) * m.splitRatio)
\tif listWidth < 20 {
\t\tlistWidth = 20
\t}
\tdetailWidth := m.width - listWidth - 4
\tif detailWidth < 20 {
\t\tdetailWidth = 20
\t}

\tm.list.SetSize(listWidth-2, mainHeight-2)
\tm.viewport.Width = detailWidth - 2
\tm.viewport.Height = mainHeight - 2
\tm.help.Width = m.width
}
''',
        '''func (m *MainModel) recalcPanelSizes() {
\tm.help.Width = m.width
\thelpHeight := 2
\tif m.help.ShowAll {
\t\thelpHeight = lipgloss.Height(m.help.View(keys))
\t\tif helpHeight < 1 {
\t\t\thelpHeight = 1
\t\t}
\t}
\tfilterBarHeight := 0
\tif m.list.FilterState() == list.Filtering {
\t\tfilterBarHeight = 1
\t}
\tstatusBarHeight := 1
\tmainHeight := m.height - helpHeight - filterBarHeight - statusBarHeight
\tif mainHeight < 3 {
\t\tmainHeight = 3
\t}

\tlistWidth := int(float64(m.width) * m.splitRatio)
\tif listWidth < 20 {
\t\tlistWidth = 20
\t}
\tdetailWidth := m.width - listWidth - 4
\tif detailWidth < 20 {
\t\tdetailWidth = 20
\t}

\tm.list.SetSize(listWidth-2, mainHeight-2)
\tm.viewport.Width = detailWidth - 2
\tm.viewport.Height = mainHeight - 2
}
''',
        "fix panel height accounting",
    )
    text = replace_once(
        text,
        '''\tcase tea.KeyMsg:
\t\tif m.list.FilterState() == list.Filtering {
\t\t\tm.list, cmd = m.list.Update(msg)
\t\t\treturn m, cmd
\t\t}
''',
        '''\tcase tea.KeyMsg:
\t\tif m.list.FilterState() == list.Filtering {
\t\t\tfilterWasVisible := true
\t\t\tm.list, cmd = m.list.Update(msg)
\t\t\tif filterWasVisible != (m.list.FilterState() == list.Filtering) {
\t\t\t\tm.recalcPanelSizes()
\t\t\t}
\t\t\treturn m, cmd
\t\t}
''',
        "resize when filter closes",
    )
    text = replace_once(
        text,
        '''\t\tcase key.Matches(msg, keys.Help):
\t\t\tm.help.ShowAll = !m.help.ShowAll
\t\t\treturn m, nil
''',
        '''\t\tcase key.Matches(msg, keys.Help):
\t\t\tm.help.ShowAll = !m.help.ShowAll
\t\t\tm.recalcPanelSizes()
\t\t\treturn m, nil
''',
        "resize for expanded help",
    )
    text = replace_once(
        text,
        '''\t\t\t\tm.list, cmd = m.list.Update(msg)
\t\t\t\tcmds = append(cmds, cmd)

\t\t\t\tif svc := m.getSelectedService(); svc != nil {
''',
        '''\t\t\t\tfilterWasVisible := m.list.FilterState() == list.Filtering
\t\t\t\tm.list, cmd = m.list.Update(msg)
\t\t\t\tif filterWasVisible != (m.list.FilterState() == list.Filtering) {
\t\t\t\t\tm.recalcPanelSizes()
\t\t\t\t}
\t\t\t\tcmds = append(cmds, cmd)

\t\t\t\tif svc := m.getSelectedService(); svc != nil {
''',
        "resize when filter opens",
    )
    text = replace_once(
        text,
        '''\tcase []client.Service:
\t\tm.services = msg
\t\tuserCount := 0
\t\tfor _, svc := range m.services {
\t\t\tif svc.Source == client.SourceUser {
\t\t\t\tuserCount++
\t\t\t}
\t\t}
\t\tm.statusMessage = fmt.Sprintf("Loaded %d services (%d user-created)", len(msg), userCount)

\t\tfiltered := m.filteredServices()
\t\tfilteredCount := len(filtered)
\t\ttotalCount := len(m.services)
\t\tif filteredCount != totalCount {
\t\t\tm.list.Title = fmt.Sprintf("User Services (%d/%d)", filteredCount, totalCount)
\t\t} else {
\t\t\tm.list.Title = fmt.Sprintf("User Services (%d)", totalCount)
\t\t}

\t\titems := m.buildListItems()
\t\tcmds = append(cmds, m.replaceListItems(items, true))
''',
        '''\tcase serviceListMsg:
\t\tif msg.id != m.refreshRequestID {
\t\t\tbreak
\t\t}
\t\tm.refreshPending = false
\t\tif msg.err != nil {
\t\t\tm.statusMessage = "Failed to list services: " + renderUserError(msg.err)
\t\t\tbreak
\t\t}
\t\tcmds = append(cmds, m.applyServiceList(msg.services))

\tcase []client.Service:
\t\t// Kept for direct model tests and callers that already have a complete list.
\t\tcmds = append(cmds, m.applyServiceList(msg))
''',
        "handle scoped service results",
    )
    if text.count('cmds = append(cmds, m.fetchServices)') != 3:
        raise RuntimeError(f"force refresh callsites: expected 3, found {text.count('cmds = append(cmds, m.fetchServices)')}")
    text = text.replace('cmds = append(cmds, m.fetchServices)', 'cmds = append(cmds, m.requestServices(true))')
    text = replace_once(
        text,
        '''\tcase followStartedMsg:
\t\tm.followPending = false
\t\tif msg.id != m.followSessionID {
\t\t\t// Stale session (selection changed while follow was starting); cancel stream.
\t\t\tmsg.cancel()
\t\t\tbreak
\t\t}
\t\tm.following = true
\t\tm.followCancel = msg.cancel
\t\tm.followLogChan = msg.ch
\t\tm.followLogLines = nil
\t\tm.followTrimmed = false
\t\tm.selectedSvc = msg.unit
\t\tcmds = append(cmds, m.continueFollow(msg.unit, msg.id))

\tcase followStoppedMsg:
\t\tif m.following && msg.id == m.followSessionID && msg.unit == m.selectedSvc {
\t\t\tm.stopFollow()
\t\t\tm.statusMessage = "Stopped following logs"
\t\t\tif m.selectedSvc != "" {
\t\t\t\tcmds = append(cmds, m.fetchDetailContent(m.selectedSvc))
\t\t\t}
\t\t}

\tcase errMsg:
\t\tm.followPending = false
\t\tm.statusMessage = fmt.Sprintf("Error: %s - %s", msg.op, renderUserError(msg.err))

\tcase tickMsg:
\t\t// Don't refresh while a filter is active or being entered — it would
\t\t// reset either the in-progress filter input or the applied filter result.
\t\tif m.list.FilterState() != list.Unfiltered {
\t\t\treturn m, m.tick()
\t\t}
\t\treturn m, tea.Batch(m.fetchServices, m.tick())
''',
        '''\tcase followStartedMsg:
\t\tif msg.id != m.followSessionID {
\t\t\t// Stale session (selection changed while follow was starting); cancel stream.
\t\t\tmsg.cancel()
\t\t\tbreak
\t\t}
\t\tm.followPending = false
\t\tm.following = true
\t\tm.followCancel = msg.cancel
\t\tm.followLogChan = msg.ch
\t\tm.followLogLines = nil
\t\tm.followTrimmed = false
\t\tm.selectedSvc = msg.unit
\t\tcmds = append(cmds, m.continueFollow(msg.unit, msg.id))

\tcase followStartFailedMsg:
\t\tif msg.id != m.followSessionID {
\t\t\tbreak
\t\t}
\t\tm.followPending = false
\t\tm.statusMessage = fmt.Sprintf("Failed to follow logs for %s: %s", msg.unit, renderUserError(msg.err))

\tcase followStreamFailedMsg:
\t\tif m.following && msg.id == m.followSessionID && msg.unit == m.selectedSvc {
\t\t\tm.stopFollow()
\t\t\tm.statusMessage = fmt.Sprintf("Log follow failed for %s: %s", msg.unit, renderUserError(msg.err))
\t\t\tif m.selectedSvc != "" {
\t\t\t\tcmds = append(cmds, m.fetchDetailContent(m.selectedSvc))
\t\t\t}
\t\t}

\tcase followStoppedMsg:
\t\tif m.following && msg.id == m.followSessionID && msg.unit == m.selectedSvc {
\t\t\tm.stopFollow()
\t\t\tm.statusMessage = "Stopped following logs"
\t\t\tif m.selectedSvc != "" {
\t\t\t\tcmds = append(cmds, m.fetchDetailContent(m.selectedSvc))
\t\t\t}
\t\t}

\tcase tickMsg:
\t\t// Don't refresh while a filter is active or being entered — it would
\t\t// reset either the in-progress filter input or the applied filter result.
\t\tif m.list.FilterState() != list.Unfiltered {
\t\t\treturn m, m.tick()
\t\t}
\t\treturn m, tea.Batch(m.requestServices(false), m.tick())
''',
        "replace follow and tick handlers",
    )
    marker = '''func (m *MainModel) replaceListItems(items []list.Item, refreshDetail bool) tea.Cmd {
'''
    helpers = '''func (m *MainModel) requestServices(force bool) tea.Cmd {
\tif m.refreshPending && !force {
\t\treturn nil
\t}
\tm.refreshRequestID++
\tm.refreshPending = true
\treturn m.fetchServices(m.refreshRequestID)
}

func (m *MainModel) applyServiceList(services []client.Service) tea.Cmd {
\tm.services = services
\tuserCount := 0
\tfor _, svc := range m.services {
\t\tif svc.Source == client.SourceUser {
\t\t\tuserCount++
\t\t}
\t}
\tm.statusMessage = fmt.Sprintf("Loaded %d services (%d user-created)", len(services), userCount)

\tfiltered := m.filteredServices()
\tfilteredCount := len(filtered)
\ttotalCount := len(m.services)
\tif filteredCount != totalCount {
\t\tm.list.Title = fmt.Sprintf("User Services (%d/%d)", filteredCount, totalCount)
\t} else {
\t\tm.list.Title = fmt.Sprintf("User Services (%d)", totalCount)
\t}

\treturn m.replaceListItems(m.buildListItems(), true)
}

'''
    text = replace_once(text, marker, helpers + marker, "insert refresh helpers")
    return text


def update_model_test(text: str) -> str:
    return replace_once(
        text,
        '''func (m *mockServiceClient) FollowLogs(ctx context.Context, name string, opts client.LogOptions) (<-chan string, context.CancelFunc, error) {
\tm.followOpts = opts
\tch := make(chan string)
\tcancel := func() { close(ch) }
\treturn ch, cancel, nil
}
''',
        '''func (m *mockServiceClient) FollowLogs(ctx context.Context, name string, opts client.LogOptions) (<-chan client.LogEvent, context.CancelFunc, error) {
\tm.followOpts = opts
\tch := make(chan client.LogEvent)
\tcancel := func() { close(ch) }
\treturn ch, cancel, nil
}
''',
        "update UI mock FollowLogs",
    )


def update_service_test(text: str) -> str:
    return replace_once(
        text,
        '''func (m *mockSystemdClient) FollowLogs(ctx context.Context, name string, opts client.LogOptions) (<-chan string, context.CancelFunc, error) {
\tif name == "error.service" {
\t\treturn nil, nil, errors.New("failed to follow logs")
\t}
\tch := make(chan string)
\tcancel := func() { close(ch) }
\treturn ch, cancel, nil
}
''',
        '''func (m *mockSystemdClient) FollowLogs(ctx context.Context, name string, opts client.LogOptions) (<-chan client.LogEvent, context.CancelFunc, error) {
\tif name == "error.service" {
\t\treturn nil, nil, errors.New("failed to follow logs")
\t}
\tch := make(chan client.LogEvent)
\tcancel := func() { close(ch) }
\treturn ch, cancel, nil
}
''',
        "update service mock FollowLogs",
    )


service_reliability_tests = r'''package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"systemd-tui/internal/client"
)

func installFakeCommand(t *testing.T, name, body string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nset -eu\n"+body), 0755); err != nil {
		t.Fatalf("write fake %s: %v", name, err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestListServicesIncludesInstalledUnloadedServices(t *testing.T) {
	userConfigDir := t.TempDir()
	userServiceDir := filepath.Join(userConfigDir, "systemd", "user")
	if err := os.MkdirAll(userServiceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userServiceDir, "created.service"), []byte("[Service]\nExecStart=/bin/true\n"), 0644); err != nil {
		t.Fatal(err)
	}

	installFakeCommand(t, "systemctl", `
case "$*" in
  *"list-unit-files"*)
    cat <<'JSON'
[
  {"unit_file":"loaded.service","state":"enabled","preset":null},
  {"unit_file":"created.service","state":"disabled","preset":null},
  {"unit_file":"package.service","state":"static","preset":null},
  {"unit_file":"worker@.service","state":"disabled","preset":null}
]
JSON
    ;;
  *"list-units"*)
    cat <<'JSON'
[
  {"unit":"loaded.service","load":"loaded","active":"active","sub":"running","description":"Loaded service"},
  {"unit":"transient.service","load":"loaded","active":"active","sub":"running","description":"Transient service"}
]
JSON
    ;;
  *)
    echo "unexpected systemctl args: $*" >&2
    exit 64
    ;;
esac
`)

	c := &systemdClient{userConfigDir: userConfigDir}
	services, err := c.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}

	if len(services) != 4 {
		t.Fatalf("got %d services, want 4: %#v", len(services), services)
	}
	byName := make(map[string]client.Service, len(services))
	for i, service := range services {
		byName[service.Name] = service
		if i > 0 && services[i-1].Name > service.Name {
			t.Fatalf("services are not sorted: %q before %q", services[i-1].Name, service.Name)
		}
	}

	loaded := byName["loaded.service"]
	if loaded.Status != client.StatusActive || loaded.Sub != "running" || loaded.Load != "loaded" || !loaded.Enabled {
		t.Fatalf("loaded runtime state was not preserved: %#v", loaded)
	}
	created := byName["created.service"]
	if created.Status != client.StatusInactive || created.Sub != "dead" || created.Load != "unloaded" || created.Source != client.SourceUser {
		t.Fatalf("installed user service fallback is wrong: %#v", created)
	}
	if packageService := byName["package.service"]; packageService.Source != client.SourceStatic {
		t.Fatalf("static service source = %q, want %q", packageService.Source, client.SourceStatic)
	}
	if _, ok := byName["transient.service"]; !ok {
		t.Fatal("loaded service without a unit-file record was dropped")
	}
	if _, ok := byName["worker@.service"]; ok {
		t.Fatal("template unit was exposed as an actionable service")
	}
}

func TestListServicesReturnsUnitFileEnumerationError(t *testing.T) {
	installFakeCommand(t, "systemctl", `
case "$*" in
  *"list-unit-files"*)
    echo "unit-file enumeration denied" >&2
    exit 7
    ;;
  *"list-units"*)
    printf '[]\n'
    ;;
esac
`)

	c := &systemdClient{}
	_, err := c.ListServices(context.Background())
	if err == nil {
		t.Fatal("expected list-unit-files failure")
	}
	if !strings.Contains(err.Error(), "list-unit-files") || !strings.Contains(err.Error(), "enumeration denied") {
		t.Fatalf("error lost command context: %v", err)
	}
}

func TestFollowLogsSurfacesProcessFailure(t *testing.T) {
	installFakeCommand(t, "journalctl", `
printf 'first line\n'
printf 'journal exploded\n' >&2
exit 42
`)

	c := &systemdClient{}
	events, cancel, err := c.FollowLogs(context.Background(), "demo.service", client.LogOptions{Lines: 10})
	if err != nil {
		t.Fatalf("FollowLogs start: %v", err)
	}
	defer cancel()

	var gotLine bool
	var streamErr error
	for event := range events {
		if event.Line == "first line" {
			gotLine = true
		}
		if event.Err != nil {
			streamErr = event.Err
		}
	}
	if !gotLine {
		t.Fatal("follow stream dropped stdout before the process failed")
	}
	if streamErr == nil {
		t.Fatal("unexpected journalctl exit was reported as a normal close")
	}
	if !strings.Contains(streamErr.Error(), "journal exploded") || !strings.Contains(streamErr.Error(), "exit status 42") {
		t.Fatalf("stream error lost stderr/exit status: %v", streamErr)
	}
}

func TestFollowLogsCancellationDoesNotReportFailure(t *testing.T) {
	installFakeCommand(t, "journalctl", `
printf 'ready\n'
while :; do sleep 1; done
`)

	c := &systemdClient{}
	events, cancel, err := c.FollowLogs(context.Background(), "demo.service", client.LogOptions{})
	if err != nil {
		t.Fatalf("FollowLogs start: %v", err)
	}

	select {
	case event := <-events:
		if event.Line != "ready" || event.Err != nil {
			t.Fatalf("unexpected first event: %#v", event)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first follow event")
	}

	cancel()
	deadline := time.After(4 * time.Second)
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			if event.Err != nil {
				t.Fatalf("explicit cancellation was surfaced as failure: %v", event.Err)
			}
		case <-deadline:
			t.Fatal("follow channel did not close after cancellation")
		}
	}
}
'''


ui_reliability_tests = r'''package ui

import (
	"errors"
	"strings"
	"testing"

	"systemd-tui/internal/client"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestStaleServiceRefreshResultIsIgnored(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{{Name: "current.service", Status: client.StatusActive}})
	model.refreshRequestID = 5
	model.refreshPending = true

	updated, _ := model.Update(serviceListMsg{
		id:       4,
		services: []client.Service{{Name: "stale.service", Status: client.StatusInactive}},
	})
	model = updated.(MainModel)

	if len(model.services) != 1 || model.services[0].Name != "current.service" {
		t.Fatalf("stale refresh overwrote current data: %#v", model.services)
	}
	if !model.refreshPending {
		t.Fatal("stale refresh cleared the current request's pending state")
	}
}

func TestLatestServiceRefreshAppliesAndClearsPending(t *testing.T) {
	model := newTestModel()
	model.refreshRequestID = 5
	model.refreshPending = true

	updated, _ := model.Update(serviceListMsg{
		id:       5,
		services: []client.Service{{Name: "latest.service", Status: client.StatusActive}},
	})
	model = updated.(MainModel)

	if model.refreshPending {
		t.Fatal("latest refresh did not clear pending state")
	}
	if len(model.services) != 1 || model.services[0].Name != "latest.service" {
		t.Fatalf("latest refresh was not applied: %#v", model.services)
	}
}

func TestServiceRefreshErrorPreservesCurrentList(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{{Name: "current.service", Status: client.StatusActive}})
	model.refreshRequestID = 3
	model.refreshPending = true

	updated, _ := model.Update(serviceListMsg{id: 3, err: errors.New("bus unavailable")})
	model = updated.(MainModel)

	if model.refreshPending {
		t.Fatal("failed current refresh left pending set")
	}
	if len(model.services) != 1 || model.services[0].Name != "current.service" {
		t.Fatalf("failed refresh destroyed the current list: %#v", model.services)
	}
	if !strings.Contains(model.statusMessage, "bus unavailable") {
		t.Fatalf("refresh error was not surfaced: %q", model.statusMessage)
	}
}

func TestPeriodicRefreshDoesNotOverlapButForcedRefreshSupersedes(t *testing.T) {
	model := newTestModel()
	model.refreshRequestID = 8
	model.refreshPending = true

	if cmd := model.requestServices(false); cmd != nil {
		t.Fatal("periodic refresh overlapped an in-flight request")
	}
	cmd := model.requestServices(true)
	if cmd == nil {
		t.Fatal("forced post-action refresh did not supersede the in-flight request")
	}
	if model.refreshRequestID != 9 || !model.refreshPending {
		t.Fatalf("forced refresh state = id %d pending %v", model.refreshRequestID, model.refreshPending)
	}
	msg, ok := cmd().(serviceListMsg)
	if !ok || msg.id != 9 {
		t.Fatalf("forced refresh returned %#v", msg)
	}
}

func TestStaleFollowStartDoesNotClearCurrentPending(t *testing.T) {
	model := newTestModel()
	model.followSessionID = 2
	model.followPending = true
	cancelled := false
	ch := make(chan client.LogEvent)

	updated, _ := model.Update(followStartedMsg{
		id:     1,
		unit:   "old.service",
		ch:     ch,
		cancel: func() { cancelled = true },
	})
	model = updated.(MainModel)

	if !cancelled {
		t.Fatal("stale started stream was not cancelled")
	}
	if !model.followPending {
		t.Fatal("stale started message cleared the current follow request")
	}
	if model.following {
		t.Fatal("stale follow session became active")
	}
}

func TestFollowStartFailureIsScopedToSession(t *testing.T) {
	model := newTestModel()
	model.followSessionID = 4
	model.followPending = true
	model.statusMessage = "starting current"

	updated, _ := model.Update(followStartFailedMsg{id: 3, unit: "old.service", err: errors.New("old failure")})
	model = updated.(MainModel)
	if !model.followPending || model.statusMessage != "starting current" {
		t.Fatalf("stale failure changed current follow state: pending=%v status=%q", model.followPending, model.statusMessage)
	}

	updated, _ = model.Update(followStartFailedMsg{id: 4, unit: "current.service", err: errors.New("start failed")})
	model = updated.(MainModel)
	if model.followPending {
		t.Fatal("current follow-start failure did not clear pending state")
	}
	if !strings.Contains(model.statusMessage, "current.service") || !strings.Contains(model.statusMessage, "start failed") {
		t.Fatalf("current follow-start failure was not surfaced: %q", model.statusMessage)
	}
}

func TestFollowStreamFailureStopsCurrentSession(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{{Name: "demo.service", Status: client.StatusActive}})
	model.followSessionID = 7
	model.following = true
	cancelled := false
	model.followCancel = func() { cancelled = true }
	model.followLogChan = make(chan client.LogEvent)

	updated, _ := model.Update(followStreamFailedMsg{id: 7, unit: "demo.service", err: errors.New("journal died")})
	model = updated.(MainModel)

	if model.following {
		t.Fatal("failed follow stream remained active")
	}
	if !cancelled {
		t.Fatal("failed follow stream was not cleaned up")
	}
	if !strings.Contains(model.statusMessage, "journal died") {
		t.Fatalf("follow failure was mislabeled as a normal stop: %q", model.statusMessage)
	}
}

func TestFilteringReservesAndReleasesTerminalRow(t *testing.T) {
	model := newTestModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(MainModel)
	baseHeight := model.viewport.Height

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model = updated.(MainModel)
	if model.list.FilterState() != list.Filtering {
		t.Fatalf("filter did not enter input state: %v", model.list.FilterState())
	}
	if model.viewport.Height != baseHeight-1 {
		t.Fatalf("viewport height while filtering = %d, want %d", model.viewport.Height, baseHeight-1)
	}
	if got := lipgloss.Height(model.View()); got > model.height {
		t.Fatalf("filter UI renders %d rows into a %d-row terminal", got, model.height)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(MainModel)
	if model.viewport.Height != baseHeight {
		t.Fatalf("viewport height after closing filter = %d, want %d", model.viewport.Height, baseHeight)
	}
}

func TestExpandedHelpAndFilterFitTerminal(t *testing.T) {
	model := newTestModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(MainModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model = updated.(MainModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model = updated.(MainModel)

	if got := lipgloss.Height(model.View()); got > model.height {
		t.Fatalf("expanded help + filter render %d rows into a %d-row terminal", got, model.height)
	}
}
'''


if "type LogEvent struct" not in Path("internal/client/types.go").read_text():
    update("internal/client/types.go", update_client_types)
    update("internal/service/client.go", update_service_client)
    update("internal/ui/commands.go", update_commands)
    update("internal/ui/model.go", update_model)
    update("internal/ui/model_test.go", update_model_test)
    update("internal/service/client_test.go", update_service_test)
    Path("internal/service/reliability_test.go").write_text(service_reliability_tests)
    Path("internal/ui/reliability_test.go").write_text(ui_reliability_tests)
else:
    print("Reliability patch already applied; leaving source unchanged.")
