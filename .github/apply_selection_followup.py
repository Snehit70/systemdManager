from pathlib import Path

path = Path("internal/ui/model.go")
text = path.read_text()

old = '''\t\tcase key.Matches(msg, keys.ToggleGroup):
\t\t\tm.groupMode = m.groupMode.Next()
\t\t\tm.statusMessage = fmt.Sprintf("Group by: %s", m.groupMode)
\t\t\tm.updateListTitle()
\t\t\tcmds = append(cmds, m.replaceListItems(m.buildListItems(), false))

\t\tcase key.Matches(msg, keys.ToggleSource):
\t\t\tm.filterMode = m.filterMode.Next()
\t\t\tm.statusMessage = fmt.Sprintf("Filter: %s", m.filterMode)
\t\t\tm.updateListTitle()
\t\t\tcmds = append(cmds, m.replaceListItems(m.buildListItems(), false))
'''

new = '''\t\tcase key.Matches(msg, keys.ToggleGroup):
\t\t\tm.groupMode = m.groupMode.Next()
\t\t\tm.statusMessage = fmt.Sprintf("Group by: %s", m.groupMode)
\t\t\tm.updateListTitle()
\t\t\tcmds = append(cmds, m.replaceListItems(m.buildListItems(), false))
\t\t\treturn m, tea.Batch(cmds...)

\t\tcase key.Matches(msg, keys.ToggleSource):
\t\t\tm.filterMode = m.filterMode.Next()
\t\t\tm.statusMessage = fmt.Sprintf("Filter: %s", m.filterMode)
\t\t\tm.updateListTitle()
\t\t\tcmds = append(cmds, m.replaceListItems(m.buildListItems(), false))
\t\t\treturn m, tea.Batch(cmds...)
'''

count = text.count(old)
if count != 1:
    raise RuntimeError(f"toggle rebuild block: expected one match, found {count}")

path.write_text(text.replace(old, new, 1))
