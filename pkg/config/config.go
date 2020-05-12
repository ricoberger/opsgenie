package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the structure of the configuration file.
type Config struct {
	ApiKey    string `yaml:"apiKey"`
	ApiUrl    string `yaml:"apiUrl"`
	User      string `yaml:"user"`
	Templates struct {
		Active   string `yaml:"active"`
		Inactive string `yaml:"inactive"`
		Selected string `yaml:"selected"`
		Details  string `yaml:"details"`
	} `yaml:"templates"`
}

// LoadConfig reads the configuration file and unmarshal the data into the config struct.
func (c *Config) LoadConfig(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	if c.Templates.Active == "" {
		c.Templates.Active = "→  {{ .Message | blue }}"
	}

	if c.Templates.Inactive == "" {
		c.Templates.Inactive = "{{ .Message | blue }}"
	}

	if c.Templates.Selected == "" {
		c.Templates.Selected = "{{ .Message | red }}"
	}

	if c.Templates.Details == "" {
		c.Templates.Details = `
--------- Details ----------
{{ "Message:" | faint }}	{{ .Message }}
{{ "Priority:" | faint }}	{{ .Priority }}
{{ "Status:" | faint }}	{{ .Status }}
{{ "Acknowledged:" | faint }}	{{ .Acknowledged }}
{{ "Tags:" | faint }}	{{ .Tags }}
{{ "Description:" | faint }}
{{ .Description }}`
	}

	return nil
}
