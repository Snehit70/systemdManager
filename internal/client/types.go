package client

import (
	"context"
	"os/exec"
	"time"
)

// ServiceStatus represents the current state of a service
type ServiceStatus string

const (
	StatusActive       ServiceStatus = "active"
	StatusInactive     ServiceStatus = "inactive"
	StatusFailed       ServiceStatus = "failed"
	StatusActivating   ServiceStatus = "activating"
	StatusDeactivating ServiceStatus = "deactivating"
	StatusReloading    ServiceStatus = "reloading"
)

// ServiceSource indicates where a service unit file originates from
type ServiceSource string

const (
	SourceUser      ServiceSource = "user"      // Created by user in ~/.config/systemd/user/
	SourceSystem    ServiceSource = "system"    // System-provided in /usr/lib/systemd/user/
	SourceGenerated ServiceSource = "generated" // Auto-generated (desktop autostart)
	SourceTransient ServiceSource = "transient" // Runtime-created
	SourceStatic    ServiceSource = "static"    // Static/alias units
	SourceUnknown   ServiceSource = "unknown"   // Unable to determine
)

// Service represents a managed systemd service
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

// LogOptions controls log retrieval behavior
type LogOptions struct {
	Lines  int    // Number of lines to fetch (default: 50)
	Follow bool   // Whether to follow/stay attached (live tailing)
	Filter string // Optional log level or text filter
}

// ServiceTemplate defines parameters for creating a new service
type ServiceTemplate struct {
	Name             string
	Description      string
	ExecStart        string
	WorkingDirectory string
	Type             string // simple, oneshot, forking, etc.
	Restart          string // on-failure, always, no
}

// ServiceClient defines the interface for service management backends
type ServiceClient interface {
	// List all available services
	ListServices() ([]Service, error)

	// Service lifecycle actions
	StartService(name string) error
	StopService(name string) error
	RestartService(name string) error
	EnableService(name string) error
	DisableService(name string) error

	// Service information
	GetStatus(name string) (ServiceStatus, error)
	GetLogs(name string, opts LogOptions) (string, error)
	GetConfig(name string) (string, error)

	// Streaming logs - returns a channel that emits log lines
	// Call cancel() to stop the stream
	FollowLogs(name string, opts LogOptions) (<-chan string, context.CancelFunc, error)

	// Service file management
	EditService(name string) (*exec.Cmd, error)
	ReloadDaemon() error

	// Service creation
	CreateService(template ServiceTemplate) error
}
