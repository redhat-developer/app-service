package oc_client

import (
	"fmt"
	"testing"
)

func TestGetComponentDesc(t *testing.T) {
	client := OcClient{}
	c, err := client.GetComponentDesc("f8-ui")
	if err != nil {
		t.Fail()
	}
	fmt.Println(string(c))
}
