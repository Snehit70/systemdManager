package client

import (
	"time"

	"os/exec"
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

// Service represents a managed systemd service
type Service struct {
	Name        string
	Description string
	Status      ServiceStatus
	Sub         string    // Sub-state (e.g., "running", "dead")
	Enabled     bool      // Whether service starts on boot
	Load        string    // Load state (e.g., "loaded", "not-found")
	PID         int       // Process ID if running
	Memory      int64     // Memory usage in bytes
	CPU         float64   // CPU usage percentage
	Since       time.Time // When service entered current status
}

// LogOptions controls log retrieval behavior
type LogOptions struct {
	Lines  int    // Number of lines to fetch (default: 50)
	Follow bool   // Whether to follow/stay attached (live tailing)
	Filter string // Optional log level or text filter
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

	// Service file management
	EditService(name string) (*exec.Cmd, error)
	ReloadDaemon() error
}
