package che_devfile

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// Tools define Che plugin
type Tools struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	ID          string `yaml:"id,omitempty"`
	Image       string `yaml:"image,omitempty"`
	MemoryLimit string `yaml:"memoryLimit,omitempty"`
}

// Source code information
type Source struct {
	Type     string `yaml:"type"`
	Location string `yaml:"location"`
}

// Project defines project information which would get imported in the che workspace.
type Project struct {
	Name   string `yaml:"name"`
	Source Source `yaml:"source"`
}

// Commands would have list of required actions(eg. run, debug etc)
type Commands struct {
	Name    string   `yaml:"name"`
	Actions []Action `yaml:"actions"`
}

// Action defines commands
type Action struct {
	Type    string `yaml:"type"`
	Tool    string `yaml:"tool"`
	Command string `yaml:"command"`
	Workdir string `yaml:"workdir"`
}

// CheDevfile root structure
type CheDevfile struct {
	SpecVersion string     `yaml:"specVersion"`
	Name        string     `yaml:"name"`
	Projects    []Project  `yaml:"projects"`
	Tools       []Tools    `yaml:"tools"`
	Commands    []Commands `yaml:"commands,omitempty"`
}

// GetDefaultCheFile returns default devfile with basic plugins.
func GetDefaultCheFile(name string) *CheDevfile {
	c := CheDevfile{
		SpecVersion: "0.0.1",
		Name:        name,
		Tools: []Tools{
			Tools{Name: "theia-editor", Type: "cheEditor", ID: "org.eclipse.che.editor.theia:master"},
			Tools{Name: "exec-plugin", Type: "chePlugin", ID: "che-machine-exec-plugin:0.0.1"},
		},
	}
	return &c
}

// AddProject expects project information to be passed to add under workspace
func (c *CheDevfile) AddProject(name, typ, location string) {
	project := Project{}
	project.Name = name
	project.Source = Source{Type: typ, Location: location}
	c.Projects = append(c.Projects, project)
}

// AddCommand adds command to devfile
func (c *CheDevfile) AddCommand(name, t, tool, command, workdir string) {
	cc := Commands{
		Name: name,
		Actions: []Action{
			Action{
				Type:    t,
				Tool:    tool,
				Command: command,
				Workdir: workdir,
			},
		},
	}
	c.Commands = append(c.Commands, cc)
}

// AddLanugagePlugins attach necessery plugins based on build type
func (c *CheDevfile) AddLanugagePlugins(buildType string) error {
	tools, ok := LANGUAGE_PLUGINS[buildType]
	if ok {
		c.Tools = append(c.Tools, tools...)
		return nil
	}
	return fmt.Errorf("Buildtype %s is not supported", buildType)
}

// String convert checfile to yaml string
func (c *CheDevfile) String() string {
	devFile, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return string(devFile)
}
