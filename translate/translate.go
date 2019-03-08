package translate

import (
	"github.com/component-to-devfile/che_devfile"
	"gopkg.in/yaml.v2"
)

type Translater struct {
	Input []byte
	Type  string
}

func (t Translater) Convert() (*che_devfile.CheDevfile, error) {
	parsedComponent := ComponentCr{}
	err := yaml.Unmarshal(t.Input, &parsedComponent)
	if err != nil {
		return nil, err
	}
	devFile := che_devfile.GetDefaultCheFile(parsedComponent.Metadata.Name)
	devFile.AddProject(
		parsedComponent.Metadata.Name,
		"git",
		parsedComponent.Spec.Codebase,
	)
	err = devFile.AddLanugagePlugins(parsedComponent.Spec.Buildtype)
	if err != nil {
		return nil, err
	}
	return devFile, nil
}
