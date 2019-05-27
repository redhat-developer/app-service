package topology

import (
	"encoding/json"
	"fmt"
)

type Graph struct {
	Nodes  []node  `json:"nodes,omitempty" protobuf:"bytes,1,opt,name=nodes"`
	Edges  []edge  `json:"edges,name=edges"`
	Groups []group `json:"groups,name=groups"`
}

type node struct {
	Id   string `json:"id,name=id"`
	Name string `json:"name,name=name"`
}

type edge struct {
	Source string `json:"source, name=source"`
	Target string `json:"target, name=target"`
}

type group struct {
	Id    string   `json:"id, name=id"`
	Name  string   `json:"name, name=name"`
	Nodes []string `json:"nodes, name=nodes"`
}

type Resource struct {
	Metadata string `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Status   string `json:"status,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Name     string `json:"name,name=name"`
	Kind     string `json:"kind,name=kind"`
}

type NodeData struct {
	NodeType  string     `json:"nodeType,name=nodeType"`
	Id        string     `json:"id,name=id"`
	Resources []Resource `json:"resource,omitempty" protobuf:"bytes,1,opt,name=resource"`
	Data      data       `json:"data, name=data"`
}

type NodeID string

type data struct {
	Url          string            `json:"url, name=url"`
	EditUrl      string            `json:"editUrl, name=editUrl"`
	BuilderImage string            `json:"builderImage, name=builderImage"`
	DonutStatus  map[string]string `json:"donutStatus, name=donutStatus"`
}

type Topology map[NodeID]NodeData

type serverMetadata struct {
	Commit string `json:"commit"`
}

type VisualizationResponse struct {
	Graph          `json:"graph"`
	Topology       `json:"topology"`
	serverMetadata `json:"serverData"`
}

func GetSampleTopology(nodes []string, resources map[string]string, groups []string, edges []string) VisualizationResponse {
	var allNodes []node
	for _, elem := range nodes {
		dataNode := &node{}
		err := json.Unmarshal([]byte(elem), dataNode)
		fmt.Println(err)
		allNodes = append(allNodes, *dataNode)
	}

	m := make(map[NodeID]NodeData)
	for k, elem := range resources {
		dataResource := &NodeData{}
		err := json.Unmarshal([]byte(elem), dataResource)
		fmt.Println(err)
		var key NodeID
		key = NodeID(k)
		m[key] = *dataResource
	}

	var allGroups []group
	for _, elem := range groups {
		dataNode := &group{}
		err := json.Unmarshal([]byte(elem), dataNode)
		fmt.Println(err)
		allGroups = append(allGroups, *dataNode)
	}

	var allEdges []edge
	for _, elem := range edges {
		dataNode := &edge{}
		err := json.Unmarshal([]byte(elem), dataNode)
		fmt.Println(err)
		allEdges = append(allEdges, *dataNode)
	}

	return VisualizationResponse{
		Graph:          Graph{Nodes: allNodes, Groups: allGroups, Edges: allEdges},
		Topology:       m,
		serverMetadata: serverMetadata{Commit: "commit"},
	}
}
