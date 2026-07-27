package client

import (
	"context"
	"os/exec"
	"time"
)

// ServiceStatus represents the current state of a service.
type ServiceStatus string

const (
	StatusActive       ServiceStatus = "active"
	StatusInactive     ServiceStatus = "inactive"
	StatusFailed       ServiceStatus = "failed"
	StatusActivating   ServiceStatus = "activating"
	StatusDeactivating ServiceStatus = "deactivating"
	StatusReloading    ServiceStatus = "reloading"
)

// ServiceSource indicates where a service unit file originates from.
type ServiceSource string

const (
	SourceUser      ServiceSource = "user"      // Created by user in ~/.config/systemd/user/
	SourceSystem    ServiceSource = "system"    // System-provided in /usr/lib/systemd/user/
	SourceGenerated ServiceSource = "generated" // Auto-generated (desktop autostart)
	SourceTransient ServiceSource = "transient" // Runtime-created
	SourceStatic    ServiceSource = "static"    // Static/alias units
	SourceUnknown   ServiceSource = "unknown"   // Unable to determine
)

// Service represents a managed systemd service.
type Service struct {
	Name        string
	Description string
	Status      ServiceStatus
	Sub         string        // Sub-state (e.g., "running", "dead")
	Enabled     bool          // Whether service starts on boot
	Load        string        // Load state (e.g., "loaded", "not-found")
	Source      ServiceSource // Where the unit file originates
	PID         int           // Process ID if running
	Memory      int64         // Memory usage in bytes
	CPU         float64       // CPU usage percentage
	Since       time.Time     // When service entered current status
}

// LogOptions controls log retrieval behavior.
type LogOptions struct {
	Lines  int    // Number of lines to fetch (default: 50)
	Follow bool   // Whether to follow/stay attached (live tailing)
	Filter string // Optional log level or text filter
}

// LogEvent carries either one live log line or a terminal stream error.
type LogEvent struct {
	Line string
	Err  error
}

// ServiceTemplate defines parameters for creating a new service.
type ServiceTemplate struct {
	Name             string
	Description      string
	ExecStart        string
	WorkingDirectory string
	Type             string // simple, oneshot, forking, etc.
	Restart          string // on-failure, always, no
}

// ServiceClient defines the interface for service management backends.
// All blocking operations accept context.Context for cancellation support.
type ServiceClient interface {
	// ListServices retrieves all available services.
	ListServices(ctx context.Context) ([]Service, error)

	// Service lifecycle actions
	StartService(ctx context.Context, name string) error
	StopService(ctx context.Context, name string) error
	RestartService(ctx context.Context, name string) error
	EnableService(ctx context.Context, name string) error
	DisableService(ctx context.Context, name string) error

	// Service information
	GetStatus(ctx context.Context, name string) (ServiceStatus, error)
	GetLogs(ctx context.Context, name string, opts LogOptions) (string, error)
	GetConfig(ctx context.Context, name string) (string, error)
	GetStatusDetails(ctx context.Context, name string) (string, error)

	// FollowLogs returns a channel that emits log lines and terminal stream errors.
	// Call the returned cancel function to stop the stream; cancellation is not emitted as an error.
	FollowLogs(ctx context.Context, name string, opts LogOptions) (<-chan LogEvent, context.CancelFunc, error)

	// Service file management
	EditService(ctx context.Context, name string) (*exec.Cmd, error)
	ReloadDaemon(ctx context.Context) error

	// Service creation
	CreateService(ctx context.Context, template ServiceTemplate) error
}
