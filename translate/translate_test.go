package translate

import (
	"fmt"
	"testing"
)

var TEST_CR = `
apiVersion: devopsconsole.openshift.io/v1alpha1
kind: Component
metadata:
  name: wit
spec:
  app: "openshiftio"
  buildtype: "go"
  codebase: "https://github.com/sbose78/nodejs-rest-http-crud"
  listenport: 8080
`

func TestNewTranslater(t *testing.T) {
	translater := Translater{
		Input: TEST_CR,
	}
	devfile := translater.Convert()
	fmt.Println("Devfile", devfile.String())
}
