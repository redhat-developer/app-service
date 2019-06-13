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
	ID          string
	Name        string
	Type        string
	Kind        string
	Value       interface{}
	Labels      map[string]string
	Annotations map[string]string
}

type innerData struct {
	nm nodeMeta
	nd topology.NodeData
}

type nodesMap struct {
	nodes map[string]innerData
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleTopology returns the handler function for the /status endpoint.
func (srv *AppServer) HandleTopology() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Get the namespace and convert the connection to a web stocket.
		namespace := r.FormValue("namespace")
		ws := convertHTTPToWebSocket(w, r)
		createTopology(ws, namespace)
	}
}

// Create and stream topology.
func createTopology(ws *websocket.Conn, namespace string) {

	// Create a client.
	k := kubeclient.NewKubeClient()

	// Create mutex.
	mutex := &sync.Mutex{}

	// Create a node watcher.
	newWatch := createNodeWatcher(namespace, k)

	var nMap nodesMap
	nMap.nodes = make(map[string]innerData)
	resourceWatchers := make(map[*watcher.Watch]nodeMeta)
	go func() {
		newWatch.ListenWatcher(func(event watch.Event) {
			node := getNodeMetadata(event.Object)
			// If event type was "deleted", delete the node. Otherwise,
			// add or update the node.
			if event.Type == "DELETED" {
				nMap.deleteNode(node)
			} else {
				nMap.addOrUpdateNode(node)
			}

			// Get all the list options for each unique node name.
			resourceOptsList := getResourcesListOptions(nMap.getLabelData("app.kubernetes.io/name", ""))

			// Create and start all watchers for list options.
			for opts, nm := range resourceOptsList {
				resourceWatchers[createResourceWatcher(namespace, opts, k)] = nm
			}

			// Listen to each resource watcher.
			for w, rwNode := range resourceWatchers {
				go func(rwNode nodeMeta, w *watcher.Watch) {
					w.ListenWatcher(func(resourceEvent watch.Event) {

						// Get the resource.
						r := getResource(resourceEvent.Object)
						if resourceEvent.Type == "DELETED" {

							// If the event  type was "deleted" delete the resource.
							nMap.deleteNodeResource(rwNode, r)
						} else {
							// If the event was to add create and add it.
							item := nMap.nodes[rwNode.Name].nd
							if item.ID == "" {
								var resource []topology.Resource
								resource = append(resource, getResource(rwNode.Value))
								item = topology.NodeData{
									Resources: resource,
									ID:        rwNode.ID,
									Type:      rwNode.Type,
									Data: topology.Data{
										URL:          "dummy_url",
										EditURL:      "dummy_edit_url",
										BuilderImage: rwNode.Name,
										DonutStatus:  make(map[string]string),
									},
								}
								var iData innerData
								iData.nm = rwNode
								iData.nd = item
								nMap.nodes[rwNode.Name] = iData
							}
							// If the resource does not exist yet, add it. Otherwise,
							// update the old resource with the new one.
							nMap.addOrUpdateNodeResource(rwNode.Name, r)
						}

						// Write the topology.
						mutex.Lock()
						ws.WriteJSON(topology.GetSampleTopology(nMap.getNode(), nMap.getResources(), nMap.getGroups(), nMap.getEdges()))
						mutex.Unlock()
					})
				}(rwNode, w)
			}
		})
	}()
}

// Get topology resources.
func (nMap nodesMap) getResources() map[string]string {
	resourceMap := make(map[string]string)
	for _, iData := range nMap.nodes {
		if iData.nd.ID != "" {
			nData, err := json.Marshal(iData.nd)
			if err != nil {
				k8log.Error(err, "failed to retrieve json encoding of node")
			}
			resourceMap[iData.nd.ID] = string(nData)
		}
	}

	return resourceMap
}

// Get topology edges.
func (nMap nodesMap) getEdges() []string {
	var edges []string
	sourceObjects := make(map[string][]nodeMeta)
	targetObjects := make(map[string][]string)

	// Arrange keys and target objects.
	targetObjects = nMap.getAnnotationData("app.openshift.io/connects-to")

	// Arrange keys and source objects.
	for targetKey, _ := range targetObjects {
		sourceObjects[targetKey] = append(sourceObjects[targetKey], nMap.getLabelData("app.kubernetes.io/name", targetKey)[targetKey]...)
	}

	// Lookup the target key in the source key and
	// construct the edge.
	for targetKey, targets := range targetObjects {
		sourceObjects := sourceObjects[targetKey]

		for _, target := range targets {
			for _, source := range sourceObjects {
				e, err := json.Marshal(topology.Edge{Source: source.ID, Target: target})
				if err != nil {
					k8log.Error(err, "failed to retrieve json encoding of node")
				}
				edges = append(edges, string(e))
			}
		}
	}

	return edges
}

// Get topology groups.
func (nMap nodesMap) getGroups() []string {
	nodes := make(map[string][]nodeMeta)
	var groups []string
	var groupNodes []string

	// Get all nodes which belong to the same part-of collection.
	nodes = nMap.getLabelData("app.kubernetes.io/part-of", "")
	for groupName, nodeMetas := range nodes {
		for _, nm := range nodeMetas {
			groupNodes = append(groupNodes, nm.ID)
		}

		// Create the group.
		g, err := json.Marshal(topology.Group{ID: "group:" + groupName, Name: groupName, Nodes: groupNodes})
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of node")
		}

		// Append the group to the list of groups.
		groups = append(groups, string(g))
	}

	return groups
}

// Create topology node.
func (nMap nodesMap) getNode() []string {
	var nodes []string
	for _, node := range nMap.nodes {
		n, err := json.Marshal(topology.Node{Name: node.nm.Name, ID: node.nm.ID})
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of node")
		}
		nodes = addOrUpdateString(nodes, string(n))
	}

	return nodes
}

// Get node label data.
func (nMap nodesMap) getLabelData(label string, keyLabel string) map[string][]nodeMeta {
	labelMap := make(map[string][]nodeMeta)
	for _, node := range nMap.nodes {

		lkey := node.nm.Labels[label]
		if lkey != "" {
			if keyLabel == "" {
				labelMap[lkey] = append(labelMap[lkey], node.nm)
			} else if keyLabel == lkey {
				labelMap[lkey] = append(labelMap[lkey], node.nm)
			}
		}
	}

	return labelMap
}

// Get node annotation data.
func (nMap nodesMap) getAnnotationData(annotation string) map[string][]string {
	annotationsMap := make(map[string][]string)
	for _, node := range nMap.nodes {
		var keys []string
		err := json.Unmarshal([]byte(node.nm.Annotations[annotation]), &keys)
		if err != nil {
			k8log.Error(err, "failed to retrieve json dencoding of node")
		}
		for _, key := range keys {
			annotationsMap[key] = append(annotationsMap[key], node.nm.ID)
		}
	}

	return annotationsMap
}

// Delete entire node.
func (nMap nodesMap) deleteNode(nm nodeMeta) {
	delete(nMap.nodes, nm.Name)
}

// Compare and add if resource does not exist or update if resource does exist.
func (nMap nodesMap) addOrUpdateNode(node nodeMeta) {
	iData := nMap.nodes[node.Name]
	iData.nm = node
	nMap.nodes[node.Name] = iData
}

// Delete a single resource on node.
func (nMap nodesMap) deleteNodeResource(nm nodeMeta, r topology.Resource) {
	var newSlice []topology.Resource
	nodeData := nMap.nodes[nm.Name].nd
	for _, resource := range nodeData.Resources {
		if resource.Kind != r.Kind {
			newSlice = append(newSlice, resource)
		}
	}

	nodeData.Resources = newSlice
	iData := nMap.nodes[nm.Name]
	iData.nd = nodeData
	nMap.nodes[nm.Name] = iData
}

// Compare and add if resource does not exist or update if resource does exist.
func (nMap nodesMap) addOrUpdateNodeResource(name string, r topology.Resource) {

	for index, element := range nMap.nodes[name].nd.Resources {
		if element.Kind == r.Kind {
			nMap.nodes[name].nd.Resources[index] = r
			return
		}
	}
	node := nMap.nodes[name]
	resources := node.nd.Resources
	resources = append(resources, r)
	node.nd.Resources = resources
	nMap.nodes[name] = node
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
	newWatch.SetFilters([]watch.EventType{watch.Added, watch.Modified, watch.Deleted})
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
	newWatch.SetFilters([]watch.EventType{watch.Added, watch.Modified, watch.Deleted})
	newWatch.StartWatcher()

	return newWatch
}

// Compile the metav1.ListOptions for resources.
func getResourcesListOptions(nMetaMap map[string][]nodeMeta) map[metav1.ListOptions]nodeMeta {
	listOptions := make(map[metav1.ListOptions]nodeMeta)

	// For each node, create the metav1.ListOptions based off
	// the app.kubernetes.io/name label.
	for lKey, nMetas := range nMetaMap {
		if lKey != "" {
			options := metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", lKey),
			}
			for _, meta := range nMetas {
				listOptions[options] = meta
			}
		}
	}

	return listOptions
}

// Gets node metadata.
func getNodeMetadata(x interface{}) nodeMeta {
	var node nodeMeta
	switch x.(type) {
	case *deploymentconfigv1.DeploymentConfig:
		dc := x.(*deploymentconfigv1.DeploymentConfig)
		node = nodeMeta{
			ID:          base64.StdEncoding.EncodeToString([]byte(dc.UID)),
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
			ID:          base64.StdEncoding.EncodeToString([]byte(d.UID)),
			Name:        d.Name,
			Kind:        "Deployment",
			Type:        "workload",
			Value:       x,
			Labels:      d.Labels,
			Annotations: d.Annotations,
		}
	default:
		k8log.Info(fmt.Sprintf("failed to recognize node type: %s", reflect.TypeOf(x)))
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
		k8log.Info(fmt.Sprintf("failed to recognize resource type: %s", reflect.TypeOf(rx)))
	}

	return r
}

// Compare and add if resource does not exist or update if resource does exist.
func addOrUpdateString(slice []string, str string) []string {
	for index, element := range slice {
		if element == str {
			slice[index] = str
			return slice
		}
	}

	return append(slice, str)
}
