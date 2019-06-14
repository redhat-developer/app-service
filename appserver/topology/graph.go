package topology

import (
	"encoding/json"
)

// Graph contains the groupds, edges and nodes of the graph.
type Graph struct {
	Nodes  []Node  `json:"nodes,omitempty" protobuf:"bytes,1,opt,name=nodes"`
	Edges  []Edge  `json:"edges,name=edges"`
	Groups []Group `json:"groups,name=groups"`
}

// Node is the id and name of a node.
type Node struct {
	ID   string `json:"id,name=id"`
	Name string `json:"name,name=name"`
}

// Edge is the source and target of a node.
type Edge struct {
	Source string `json:"source,name=source"`
	Target string `json:"target,name=target"`
}

// Group is a group of nodes.
type Group struct {
	ID    string   `json:"id,name=id"`
	Name  string   `json:"name,name=name"`
	Nodes []string `json:"nodes,name=nodes"`
}

// Resource of a node.
type Resource struct {
	Metadata string `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Status   string `json:"status,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Name     string `json:"name,name=name"`
	Kind     string `json:"kind,name=kind"`
	Spec     string `json:"spec,name=spec"`
}

// NodeData is the node data.
type NodeData struct {
	Type      string     `json:"type,name=type"`
	ID        string     `json:"id,name=id"`
	Resources []Resource `json:"resource,omitempty" protobuf:"bytes,1,opt,name=resource"`
	Data      Data       `json:"data,name=data"`
}

// NodeID is the node id.
type NodeID string

// Data value
type Data struct {
	URL          string            `json:"url,name=url"`
	EditURL      string            `json:"editUrl,name=editUrl"`
	BuilderImage string            `json:"builderImage,name=builderImage"`
	DonutStatus  map[string]string `json:"donutStatus,name=donutStatus"`
}

// Topology value.
type Topology map[NodeID]NodeData

type serverMetadata struct {
	Commit string `json:"commit"`
}

// VisualizationResponse is the response of the api.
type VisualizationResponse struct {
	Graph          `json:"graph"`
	Topology       `json:"topology"`
	serverMetadata `json:"serverData"`
}

// GetSampleTopology compiles the nodes, resources, groups
// and edges to create the json VisualizationResponse.
func GetSampleTopology(nodes []string, resources map[string]string, groups []string, edges []string) VisualizationResponse {
	allNodes := []Node{}
	for _, elem := range nodes {
		dataNode := &Node{}
		err := json.Unmarshal([]byte(elem), dataNode)
		if err != nil {
			panic(err)
		}
		allNodes = append(allNodes, *dataNode)
	}

	topology := make(map[NodeID]NodeData)
	for k, elem := range resources {
		dataResource := &NodeData{}
		err := json.Unmarshal([]byte(elem), dataResource)
		if err != nil {
			panic(err)
		}
		var key NodeID
		key = NodeID(k)
		topology[key] = *dataResource
	}

	allGroups := []Group{}
	for _, elem := range groups {
		dataNode := &Group{}
		err := json.Unmarshal([]byte(elem), dataNode)
		if err != nil {
			panic(err)
		}
		allGroups = append(allGroups, *dataNode)
	}

	allEdges := []Edge{}
	for _, elem := range edges {
		dataNode := &Edge{}
		err := json.Unmarshal([]byte(elem), dataNode)
		if err != nil {
			panic(err)
		}
		allEdges = append(allEdges, *dataNode)
	}

	return VisualizationResponse{
		Graph:          Graph{Nodes: allNodes, Groups: allGroups, Edges: allEdges},
		Topology:       topology,
		serverMetadata: serverMetadata{Commit: "commit"},
	}
}
