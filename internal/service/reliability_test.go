package service

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
