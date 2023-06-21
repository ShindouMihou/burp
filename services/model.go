package services

type Dependency struct {
	Name  *string `toml:"name" json:"name"`
	Image string  `toml:"image" json:"image"`
	Container
}

type Container struct {
	Command     *string           `toml:"command" json:"command"`
	Ports       [][]uint16        `toml:"ports" json:"ports"`
	Volumes     [][]string        `toml:"volumes" json:"volumes"`
	Networks    []string          `toml:"networks" json:"networks"`
	Devices     []string          `toml:"device" json:"devices"`
	Environment map[string]string `toml:"environment" json:"environment"`
}

type Service struct {
	Name       string `toml:"name" json:"name"`
	Build      string `toml:"build" json:"build"`
	Repository string `toml:"repository" json:"repository"`
	Container
}

type Burp struct {
	Service      Service      `toml:"service" json:"service"`
	Dependencies []Dependency `toml:"dependencies" json:"dependencies"`
	Environment  *Environment `toml:"environment" json:"environment"`
}

type Environment struct {
	File         string            `toml:"file" json:"file"`
	Output       string            `toml:"output" json:"output"`
	Replacements map[string]string `toml:"replacements" json:"replacements"`
}
