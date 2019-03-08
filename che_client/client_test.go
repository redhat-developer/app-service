package che_client

import (
	"fmt"
	"testing"
)

var devFile = `
specVersion: 0.0.1
name: petclinic-dev-environment
projects:
  - name: petclinic
    source:
      type: git
      location: "https://github.com/nodeshift-starters/nodejs-health-check"
tools:
  - name: theia-ide
    type: cheEditor
    id: org.eclipse.che.editor.theia:1.0.0
  - name: terminal
    type: chePlugin
    id: che-machine-exec-plugin:0.0.1
`

func TestGetDefaultClient(t *testing.T) {
	c := CheDefaultClient()
	if c.Host == "" {
		t.Fail()
	}
}

func TestCreateWorkspace(t *testing.T) {
	c := CheDefaultClient()
	resp, _ := c.CreateWorkspace(devFile)
	fmt.Println("DATA", string(resp))
}
