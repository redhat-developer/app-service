package appserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/gorilla/websocket"
	deploymentconfigv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-developer/app-service/appserver/topology"
	"github.com/redhat-developer/app-service/kubeclient"
	"github.com/redhat-developer/app-service/watcher"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var k8log = logf.Log

type nodeMeta struct {
	Id          string
	Name        string
	Type        string
	Kind        string
	Value       interface{}
	Labels      map[string]string
	Annotations map[string]string
}

type data struct {
	nodes []nodeMeta
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleTopology returns the handler function for the /status endpoint
func (srv *AppServer) HandleTopology() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Get the namespace and convert the connection to a web stocket.
		namespace := "default"
		ws := convertHTTPToWebSocket(w, r)
		createTopology(ws, namespace)
	}
}

func createTopology(ws *websocket.Conn, namespace string) {
	// Create a client.
	k := kubeclient.NewKubeClient()

	// Create a node watcher.
	newWatch := createNodeWatcher(namespace, k)

	// Create mutex for writing data.
	mutex := &sync.Mutex{}

	// Cache topology node data.
	var nodesMap []nodeMeta
	nodeData := make(map[string]string)
	rawNodeData := make(map[string]topology.NodeData)
	var d data
	go func() {
		newWatch.ListenWatcher(func(event watch.Event) {
			node := getNodeMetadata(event)
			nodesMap = append(nodesMap, node)
			d.nodes = nodesMap

			// Get all list options for each unique node name.
			resourceOptsList := getResourcesListOptions(d.getLabelData("app.kubernetes.io/name", ""))

			// Create and start all watchers for list options
			resourceWatchers := make(map[*watcher.Watch]nodeMeta)
			for opts, v := range resourceOptsList {
				resourceWatchers[createResourceWatcher(namespace, opts, k)] = v
			}

			// Listen to each watcher.
			for w, metadata := range resourceWatchers {
				go func(metadata nodeMeta, w *watcher.Watch) {
					w.ListenWatcher(func(resourceEvent watch.Event) {
						r := getResource(resourceEvent.Object)
						item := rawNodeData[metadata.Id]
						if item.Id == "" {
							var res []topology.Resource
							res = append(res, getResource(metadata.Value))
							rawNodeData[metadata.Id] = topology.NodeData{
								Resources: res,
								Id:        metadata.Id,
								Type:      metadata.Type,
								Data: topology.Data{
									Url:          "dummy_url",
									EditUrl:      "dummy_edit_url",
									BuilderImage: metadata.Name,
									DonutStatus:  make(map[string]string),
								},
							}
							item = rawNodeData[metadata.Id]
						}

						// If the resource does not exist yet, add it. Otherwise,
						// update the old resource with the new one.
						item.Resources = addOrUpdate(item.Resources, r)
						rawNodeData[metadata.Id] = item
						nd, err := json.Marshal(item)
						if err != nil {
							k8log.Error(err, "failed to retrieve json encoding of node")
						}
						nodeData[item.Id] = string(nd)

						// Write the topology.
						mutex.Lock()
						ws.WriteJSON(topology.GetSampleTopology(d.getNode(), nodeData, d.getGroups(), d.getEdges()))
						mutex.Unlock()
					})
				}(metadata, w)
			}
		})
	}()
}

// Converts the HTTP connection to a websocket in order to stream data.
func convertHTTPToWebSocket(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}

	return ws
}

// Create a watcher for Deployments and DeploymentConfigs.
func createNodeWatcher(namespace string, k *kubeclient.KubeClient) *watcher.Watch {
	listOptions := metav1.ListOptions{}
	onGetWatchError := func(err error) {
		fmt.Errorf("Error is %+v", err)
	}
	newWatch := watcher.NewWatch(namespace,
		k,
		k.GetDeploymentConfigWatcher(namespace, listOptions, onGetWatchError),
		k.GetDeploymentWatcher(namespace, listOptions, onGetWatchError),
	)
	newWatch.SetFilters([]watch.EventType{watch.Added, watch.Modified})
	newWatch.StartWatcher()

	return newWatch
}

// Create a watcher for all resources.
func createResourceWatcher(namespace string, opts metav1.ListOptions, k *kubeclient.KubeClient) *watcher.Watch {
	onGetWatchError := func(err error) {
		fmt.Errorf("Error is %+v", err)
	}
	newWatch := watcher.NewWatch(namespace,
		k,
		k.GetDeploymentConfigWatcher(namespace, opts, onGetWatchError),
		k.GetDeploymentWatcher(namespace, opts, onGetWatchError),
		k.GetReplicationControllerWatcher(namespace, opts, onGetWatchError),
		k.GetReplicaSetWatcher(namespace, opts, onGetWatchError),
		k.GetServiceWatcher(namespace, opts, onGetWatchError),
		k.GetRouteWatcher(namespace, opts, onGetWatchError),
	)
	newWatch.SetFilters([]watch.EventType{watch.Added, watch.Modified})
	newWatch.StartWatcher()

	return newWatch
}

// Compile topology edges.
func (d data) getEdges() []string {
	var edges []string
	sourceObjects := make(map[string][]nodeMeta)
	targetObjects := make(map[string][]string)

	// Arrange keys and target objects.
	targetObjects = d.getAnnotationData("app.openshift.io/connects-to")

	// Arrange keys and source objects.
	for targetKey, _ := range targetObjects {
		sourceObjects[targetKey] = append(sourceObjects[targetKey], d.getLabelData("app.kubernetes.io/name", targetKey)[targetKey]...)
	}

	// Lookup the target key in the source key and
	// construct the edge.
	for targetKey, targets := range targetObjects {
		sourceObjects := sourceObjects[targetKey]

		for _, target := range targets {
			for _, source := range sourceObjects {
				e, err := json.Marshal(topology.Edge{Source: source.Id, Target: target})
				if err != nil {
					k8log.Error(err, "failed to retrieve json encoding of node")
				}
				edges = append(edges, string(e))
			}
		}
	}

	return edges
}

// Compile topology groups.
func (d data) getGroups() []string {
	nodes := make(map[string][]nodeMeta)
	var groups []string
	var gs []string

	// Get all nodes which belong to the same part-of collection.
	nodes = d.getLabelData("app.kubernetes.io/part-of", "")
	for key, value := range nodes {
		for _, v := range value {
			gs = append(gs, v.Id)
		}

		// Create the group.
		g, err := json.Marshal(topology.Group{Id: "group:" + key, Name: key, Nodes: gs})
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of node")
		}

		// Append the group to the list of groups.
		groups = append(groups, string(g))
	}

	return groups
}

// Compile the metav1.ListOptions for resources.
func getResourcesListOptions(dc map[string][]nodeMeta) map[metav1.ListOptions]nodeMeta {
	listOptions := make(map[metav1.ListOptions]nodeMeta)

	// For each node, create the metav1.ListOptions based off
	// the app.kubernetes.io/name label.
	for labelKey, dcNodes := range dc {
		if labelKey != "" {
			options := metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", labelKey),
			}
			for _, dc := range dcNodes {
				listOptions[options] = dc
			}
		}
	}

	return listOptions
}

// Create topology node.
func (d data) getNode() []string {
	var nodes []string
	for _, node := range d.nodes {
		n, err := json.Marshal(topology.Node{Name: node.Name, Id: node.Id})
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of node")
		}
		nodes = append(nodes, string(n))
	}

	return nodes
}

// Get node label data.
func (d data) getLabelData(label string, keyLabel string) map[string][]nodeMeta {
	metadata := make(map[string][]nodeMeta)
	for _, node := range d.nodes {
		labelValue := node.Labels[label]
		if keyLabel == "" {
			metadata[labelValue] = append(metadata[labelValue], node)
		} else if keyLabel == labelValue {
			metadata[labelValue] = append(metadata[labelValue], node)
		}
	}

	return metadata
}

// Get node annotation data.
func (d data) getAnnotationData(annotation string) map[string][]string {
	nodes := make(map[string][]string)
	for _, node := range d.nodes {
		var keys []string
		err := json.Unmarshal([]byte(node.Annotations[annotation]), &keys)
		if err != nil {
			k8log.Error(err, "failed to retrieve json dencoding of node")
		}
		for _, key := range keys {
			nodes[key] = append(nodes[key], node.Id)
		}
	}

	return nodes
}

// Gets node metadata.
func getNodeMetadata(event watch.Event) nodeMeta {
	var x interface{} = event.Object
	var node nodeMeta
	switch x.(type) {
	case *deploymentconfigv1.DeploymentConfig:
		dc := x.(*deploymentconfigv1.DeploymentConfig)
		node = nodeMeta{
			Id:          base64.StdEncoding.EncodeToString([]byte(dc.UID)),
			Name:        dc.Name,
			Kind:        "DeploymentConfig",
			Type:        "workload",
			Value:       x,
			Labels:      dc.Labels,
			Annotations: dc.Annotations,
		}
	case *appsv1.Deployment:
		d := x.(*appsv1.Deployment)
		node = nodeMeta{
			Id:          base64.StdEncoding.EncodeToString([]byte(d.UID)),
			Name:        d.Name,
			Kind:        "Deployment",
			Type:        "workload",
			Value:       x,
			Labels:      d.Labels,
			Annotations: d.Annotations,
		}
	default:
		fmt.Println(reflect.TypeOf(x))
	}

	return node
}

// Create topology resources.
func getResource(rx interface{}) topology.Resource {
	var r topology.Resource
	switch rx.(type) {
	case *deploymentconfigv1.DeploymentConfig:
		deploymentConfig := rx.(*deploymentconfigv1.DeploymentConfig)
		meta, err := json.Marshal(deploymentConfig.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of deployment config")
		}
		status, err := json.Marshal(deploymentConfig.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of deployment config")
		}
		r = topology.Resource{
			Name:     deploymentConfig.Name,
			Kind:     "DeploymentConfig",
			Metadata: string(meta),
			Status:   string(status),
		}
	case *appsv1.Deployment:
		deployment := rx.(*appsv1.Deployment)
		meta, err := json.Marshal(deployment.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of deployment")
		}
		status, err := json.Marshal(deployment.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of deployment")
		}
		r = topology.Resource{
			Name:     deployment.Name,
			Kind:     "Deployment",
			Metadata: string(meta),
			Status:   string(status),
		}
	case *corev1.Service:
		serv := rx.(*corev1.Service)
		meta, err := json.Marshal(serv.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of service")
		}
		status, err := json.Marshal(serv.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of service")
		}
		r = topology.Resource{
			Name:     serv.Name,
			Kind:     "Service",
			Metadata: string(meta),
			Status:   string(status),
		}
	case *routev1.Route:
		route := rx.(*routev1.Route)
		meta, err := json.Marshal(route.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of route")
		}
		status, err := json.Marshal(route.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of route")
		}
		r = topology.Resource{Name: route.Name,
			Kind:     "Route",
			Metadata: string(meta),
			Status:   string(status),
		}
	case *corev1.ReplicationController:
		rc := rx.(*corev1.ReplicationController)
		meta, err := json.Marshal(rc.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of replication controller")
		}
		status, err := json.Marshal(rc.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of replication controller")
		}
		r = topology.Resource{
			Name:     rc.Name,
			Kind:     "ReplicationController",
			Metadata: string(meta),
			Status:   string(status),
		}
	case *appsv1.ReplicaSet:
		rs := rx.(*appsv1.ReplicaSet)
		meta, err := json.Marshal(rs.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of replica set")
		}
		status, err := json.Marshal(rs.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of replica set")
		}
		r = topology.Resource{
			Name:     rs.Name,
			Kind:     "ReplicaSet",
			Metadata: string(meta),
			Status:   string(status)}
	default:
		fmt.Println(reflect.TypeOf(rx))
	}

	return r
}

// Compare and add if resource does not exist or update if resource does exist.
func addOrUpdate(slice []topology.Resource, i topology.Resource) []topology.Resource {
	for index, ele := range slice {
		if ele == i {
			slice[index] = i
			return slice
		}
	}

	return append(slice, i)
}
