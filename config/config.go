// Config package deals with applciation configuration
package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// App holds applciation configuration
type App struct {
	Name string `yaml:"name" json:"name"`
	Host string `yaml:"host" json:"host"`
	Port string `yaml:"port" json:"port"`
}

// DB holds database configuration
type DB struct {
	Instance string `yaml:"instance" json:"instance"`
	Port     string `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Database string `yaml:"database" json:"database"`
}

// Log holds log related configuration
type Log struct {
	Debug       bool `yaml:"debug" json:"debug"`
	JSON        bool `yaml:"json" json:"json"`
	Color       bool `yaml:"color" json:"color"`
	PrintRoutes bool `yaml:"print_routes" json:"print_routes"`
}

// Config holds application configuration
type Config struct {
	App App `yaml:"app" json:"app"`
	DB  DB  `yaml:"db" json:"db"`
	Log Log `yaml:"log" json:"log"`
}

// Read accepts multiple file paths and return last valid configuration.
// Returns error if no valid path found
func Read(cfg interface{}, paths ...string) error {
	var path string

	for _, v := range paths {
		if _, err := os.Stat(v); err == nil {
			path = v
		}
	}

	if path == "" {
		return fmt.Errorf("%w: no valid config file found", os.ErrNotExist)
	}

	return cleanenv.ReadConfig(path, cfg)
}
