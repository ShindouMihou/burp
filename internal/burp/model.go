package burp

type Service struct {
	Build      string     `toml:"build" json:"build"`
	Repository Repository `toml:"repository" json:"repository"`
	Container
}

type Repository struct {
	Link   string  `toml:"link" json:"link"`
	Branch *string `toml:"branch,omitempty"  json:"branch,omitempty"`
}

type Dependency struct {
	Image string `toml:"image" json:"image"`
	Container
}

type Container struct {
	Name             string                   `toml:"name" json:"name"`
	User             string                   `toml:"user,omitempty" json:"user,omitempty"`
	Hostname         string                   `toml:"hostname,omitempty" json:"hostname,omitempty"`
	WorkingDirectory string                   `toml:"working_directory,omitempty" json:"working_directory,omitempty"`
	Command          []string                 `toml:"command,omitempty" json:"command,omitempty"`
	Entrypoint       []string                 `toml:"entrypoint,omitempty" json:"entrypoint,omitempty"`
	Ports            [][]string               `toml:"ports,omitempty" json:"ports,omitempty"`
	Volumes          []ContainerVolume        `toml:"volumes,omitempty" json:"volumes,omitempty"`
	Limits           *ContainerResourceLimits `toml:"limits,omitempty" json:"limits,omitempty"`
	RestartPolicy    *ContainerRestartPolicy  `toml:"restart_policy,omitempty" json:"restart_policy,omitempty"`
	Networks         []string                 `toml:"networks,omitempty" json:"networks,omitempty"`
	Devices          []string                 `toml:"device,omitempty" json:"devices,omitempty"`
	DNS              []string                 `toml:"dns,omitempty" json:"dns,omitempty"`
	Environment      map[string]string        `toml:"environment,omitempty" json:"environment,omitempty"`
	Override         *bool                    `toml:"override,omitempty" json:"override"`
}

type PlatformVolume struct {
	Name     string `toml:"name" json:"name"`
	External *bool  `toml:"external,omitempty" json:"external,omitempty"`
}

type Application struct {
	Service      Service          `toml:"service" json:"service"`
	Dependencies []Dependency     `toml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Environment  *Environment     `toml:"environment,omitempty" json:"environment,omitempty"`
	Volumes      []PlatformVolume `toml:"volumes,omitempty" json:"volumes,omitempty"`
	Includes     []Include        `toml:"includes,omitempty" json:"includes,omitempty"`
}

type Environment struct {
	Baseline     string            `toml:"baseline" json:"baseline"`
	Override     bool              `toml:"override" json:"override"`
	Replacements map[string]string `toml:"replacements" json:"replacements"`
	ServerSide   bool              `toml:"server-side" json:"server-side"`
}

type Include struct {
	Source   string `toml:"source" json:"source"`
	Target   string `toml:"target" json:"target"`
	Required bool   `toml:"required,omitempty" json:"required"`
}

type HashedInclude struct {
	Include
	Hash string `toml:"hash" json:"hash"`
}
