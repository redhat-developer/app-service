package oc_client

import "os/exec"

// OcClient client to interact with oc
type OcClient struct {
}

// GetComponentDesc get component cr using oc
func (o OcClient) GetComponentDesc(name string) ([]byte, error) {
	cmd := "oc get component " + name + " -o yaml"
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return []byte(""), nil
	}
	return out, nil
}
