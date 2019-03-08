package che_devfile

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestDefaultCheFile(t *testing.T) {
	c := GetDefaultCheFile("test_project")
	c.AddLanugagePlugins("go")
	fmt.Printf("TEST %+v\n", c)
	out, _ := yaml.Marshal(c)
	fmt.Println(string(out))
}
