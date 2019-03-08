package che_client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Client struct {
	Host string
}

func CheDefaultClient() *Client {
	c := new(Client)
	c.Host = os.Getenv("CHE_HOST")
	return c
}

func (c *Client) CreateWorkspace(devfile string) ([]byte, error) {
	res, err := http.Post(
		c.Host+"/api/devfile",
		"application/yaml",
		bytes.NewBuffer([]byte(devfile)),
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Failed to create workspace %s", res.Status)
	}
	defer res.Body.Close()
	out, _ := ioutil.ReadAll(res.Body)
	return out, nil
}
