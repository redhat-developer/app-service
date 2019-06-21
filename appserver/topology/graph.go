package topology

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
	ID     string `json:"id,name=id"`
	Source string `json:"source,name=source"`
	Target string `json:"target,name=target"`
	Type   string `json:"type,name=type"`
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
	Name      string     `json:"name,name=name"`
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
func GetSampleTopology(nodes []Node, resources map[NodeID]NodeData, groups []Group, edges []Edge) VisualizationResponse {
	return VisualizationResponse{
		Graph:          Graph{Nodes: nodes, Groups: groups, Edges: edges},
		Topology:       resources,
		serverMetadata: serverMetadata{Commit: "commit"},
	}
}
