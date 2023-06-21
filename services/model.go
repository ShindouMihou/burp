package services

type Dependency struct {
	Image string `toml:"image" json:"image"`
	Container
}

type Container struct {
	Name             string                   `toml:"name" json:"name"`
	User             string                   `toml:"user" json:"user"`
	Hostname         string                   `toml:"hostname" json:"hostname"`
	WorkingDirectory string                   `toml:"working_directory" json:"working_directory"`
	Command          []string                 `toml:"command" json:"command"`
	Entrypoint       []string                 `toml:"entrypoint" json:"entrypoint"`
	Ports            [][]string               `toml:"ports" json:"ports"`
	Volumes          []ContainerVolume        `toml:"volumes" json:"volumes"`
	Limits           *ContainerResourceLimits `toml:"limits" json:"limits"`
	RestartPolicy    *ContainerRestartPolicy  `toml:"restart_policy" json:"restart_policy"`
	Networks         []string                 `toml:"networks" json:"networks"`
	Devices          []string                 `toml:"device" json:"devices"`
	DNS              []string                 `toml:"dns" json:"dns"`
	Environment      map[string]string        `toml:"environment" json:"environment"`
}

type PlatformVolume struct {
	Name     string `toml:"name" json:"name"`
	External *bool  `toml:"external" json:"external"`
}

type Service struct {
	Build      string `toml:"build" json:"build"`
	Repository string `toml:"repository" json:"repository"`
	Container
}

type Burp struct {
	Service      Service          `toml:"service" json:"service"`
	Dependencies []Dependency     `toml:"dependencies" json:"dependencies"`
	Environment  *Environment     `toml:"environment" json:"environment"`
	Volumes      []PlatformVolume `toml:"volumes" json:"volumes"`
}

type Environment struct {
	File         string            `toml:"file" json:"file"`
	Output       string            `toml:"output" json:"output"`
	Replacements map[string]string `toml:"replacements" json:"replacements"`
}
