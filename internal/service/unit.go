package service

type Unit struct {
	Unit        string `json:"unit"`
	Load        string `json:"load"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
}

type UnitFile struct {
	UnitFile string  `json:"unit_file"`
	State    string  `json:"state"`
	Preset   *string `json:"preset"`
}
