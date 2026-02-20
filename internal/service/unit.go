package service

// Unit represents a systemd unit from systemctl list-units output.
type Unit struct {
	Unit        string `json:"unit"`
	Load        string `json:"load"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
}

// UnitFile represents a systemd unit file from systemctl list-unit-files output.
type UnitFile struct {
	UnitFile string  `json:"unit_file"`
	State    string  `json:"state"`
	Preset   *string `json:"preset"`
}
