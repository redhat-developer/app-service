package appserver

import (
	//"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	"net/http"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sync"
)

var k8log = logf.Log

type nodeMeta struct {
	Id    string
	Name  string
	Type  string
	Value interface{}
}

type dataTypes struct {
	Id    string
	Key   string
	Value interface{}
}

type data struct {
	nodes map[dataTypes][]dataTypes
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleTopology returns the handler function for the /status endpoint
func (srv *AppServer) HandleTopology() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		namespace := "default"
		ws := convertHTTPToWebSocket(w, r)

		k := kubeclient.NewKubeClient()
		newWatch := createNodeWatcher(namespace, k)

		nodesMap := make(map[dataTypes][]dataTypes)
		nodeDatas := make(map[string]string)
		rawNodeData := make(map[string]topology.NodeData)
		mu := &sync.Mutex{}
		var d data
		go func() {
			newWatch.ListenWatcher(func(event watch.Event) {
				node := getNode(event)
				nodesMap[node] = append(nodesMap[node], dataTypes{})
				d.nodes = nodesMap

				resourceOptsList := getResourcesListOptions(d.getLabelData("app.kubernetes.io/name", "", true))
				resourceWatchers := make(map[nodeMeta]*watcher.Watch)
				for opts, v := range resourceOptsList {
					resourceWatchers[v] = createResourceWatcher(namespace, opts, k)
				}

				for metadata, watchh := range resourceWatchers {
					go func(metadata nodeMeta, watchh *watcher.Watch) {
						watchh.ListenWatcher(func(resourceEvent watch.Event) {
							r := getResource(resourceEvent.Object)
							item := rawNodeData[metadata.Id]
							if item.Id == "" {
								var res []topology.Resource
								res = append(res, getResource(metadata.Value))
								rawNodeData[metadata.Id] = topology.NodeData{Resources: res, Id: metadata.Id, Type: metadata.Type, Data: topology.Data{Url: "dummy_url", EditUrl: "dummy_edit_url", BuilderImage: metadata.Name, DonutStatus: make(map[string]string)}}
								item = rawNodeData[metadata.Id]
							}
							item.Resources = addOrUpdate(item.Resources, r)
							rawNodeData[metadata.Id] = item
							nd, err := json.Marshal(item)
							if err != nil {
								k8log.Error(err, "failed to retrieve json encoding of node")
							}
							nodeDatas[item.Id] = string(nd)

							mu.Lock()
							ws.WriteJSON(topology.GetSampleTopology(d.formatNodes(), nodeDatas, d.getGroups(), d.getEdges()))
							mu.Unlock()
						})
					}(metadata, watchh)
				}
			})
		}()
	}
}

func convertHTTPToWebSocket(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}

	return ws
}
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

func (d data) getEdges() []string {
	var edges []string
	sourceObjects := make(map[string][]nodeMeta)
	targetObjects := make(map[string][]string)

	// Arrange keys and target objects.
	targetObjects = d.getAnnotationData("app.openshift.io/connects-to")

	// Arrange keys and source objects.
	for targetKey, _ := range targetObjects {
		sourceObjects[targetKey] = append(sourceObjects[targetKey], d.getLabelData("app.kubernetes.io/name", targetKey, true)[targetKey]...)
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

func (d data) getGroups() []string {
	nodes := make(map[string][]nodeMeta)
	var groups []string
	var gs []string

	nodes = d.getLabelData("app.kubernetes.io/part-of", "", false)
	for key, value := range nodes {
		for _, v := range value {
			gs = append(gs, v.Id)
		}
		g, err := json.Marshal(topology.Group{Id: "group:" + key, Name: key, Nodes: gs})
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of node")
		}
		groups = append(groups, string(g))
	}

	return groups
}

func getResourcesListOptions(dc map[string][]nodeMeta) map[metav1.ListOptions]nodeMeta {
	listOptions := make(map[metav1.ListOptions]nodeMeta)
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

func formatDeploymentConfig(deploymentConfigItems interface{}) topology.Resource {
	deploymentConfig := deploymentConfigItems.(*deploymentconfigv1.DeploymentConfig)
	meta, err := json.Marshal(deploymentConfig.GetObjectMeta())
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of deployment config")
	}
	status, err := json.Marshal(deploymentConfig.Status)
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of deployment config")
	}
	return topology.Resource{Name: deploymentConfig.Name, Kind: "DeploymentConfig", Metadata: string(meta), Status: string(status)}
}

func formatDeployment(deploymentItems interface{}) topology.Resource {
	deployment := deploymentItems.(*appsv1.Deployment)
	meta, err := json.Marshal(deployment.GetObjectMeta())
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of deployment")
	}
	status, err := json.Marshal(deployment.Status)
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of deployment")
	}
	return topology.Resource{Name: deployment.Name, Kind: "Deployment", Metadata: string(meta), Status: string(status)}
}

func formatService(services interface{}) topology.Resource {
	serv := services.(*corev1.Service)
	meta, err := json.Marshal(serv.GetObjectMeta())
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of service")
	}
	status, err := json.Marshal(serv.Status)
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of service")
	}
	return topology.Resource{Name: serv.Name, Kind: "Service", Metadata: string(meta), Status: string(status)}
}

func formatReplicationController(replicationControllerItems interface{}) topology.Resource {
	rc := replicationControllerItems.(*corev1.ReplicationController)
	meta, err := json.Marshal(rc.GetObjectMeta())
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of replication controller")
	}
	status, err := json.Marshal(rc.Status)
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of replication controller")
	}
	return topology.Resource{Name: rc.Name, Kind: "ReplicationController", Metadata: string(meta), Status: string(status)}
}

func formatReplicaSet(replicaSetItems interface{}) topology.Resource {
	rs := replicaSetItems.(*appsv1.ReplicaSet)
	meta, err := json.Marshal(rs.GetObjectMeta())
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of replica set")
	}
	status, err := json.Marshal(rs.Status)
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of replica set")
	}
	return topology.Resource{Name: rs.Name, Kind: "ReplicaSet", Metadata: string(meta), Status: string(status)}
}

func formatRoute(routeItems interface{}) topology.Resource {
	r := routeItems.(*routev1.Route)
	meta, err := json.Marshal(r.GetObjectMeta())
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of route")
	}
	status, err := json.Marshal(r.Status)
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of route")
	}
	return topology.Resource{Name: r.Name, Kind: "Route", Metadata: string(meta), Status: string(status)}
}

func (d data) formatNodes() []string {
	var nodes []string

	for key, _ := range d.nodes {
		if key.Key == "DeploymentConfig" {
			dc := key.Value.(*deploymentconfigv1.DeploymentConfig)
			n, err := json.Marshal(topology.Node{Name: dc.Name, Id: base64.StdEncoding.EncodeToString([]byte(dc.UID))})
			if err != nil {
				k8log.Error(err, "failed to retrieve json encoding of node")
			}
			nodes = append(nodes, string(n))
		} else if key.Key == "Deployment" {
			d := key.Value.(*appsv1.Deployment)
			n, err := json.Marshal(topology.Node{Name: d.Name, Id: base64.StdEncoding.EncodeToString([]byte(d.UID))})
			if err != nil {
				k8log.Error(err, "failed to retrieve json encoding of node")
			}
			nodes = append(nodes, string(n))
		}
	}

	return nodes
}

func (d data) getLabelData(label string, keyLabel string, meta bool) map[string][]nodeMeta {
	nnn := make(map[string][]nodeMeta)
	for key, _ := range d.nodes {
		if key.Key == "DeploymentConfig" {
			dc := key.Value.(*deploymentconfigv1.DeploymentConfig)
			labelValue := dc.Labels[label]
			var jsn nodeMeta
			var err error
			if meta {
				jsn = nodeMeta{Name: dc.Name, Type: "workload", Id: base64.StdEncoding.EncodeToString([]byte(dc.UID)), Value: dc}
			} else {
				jsn = nodeMeta{Id: base64.StdEncoding.EncodeToString([]byte(dc.UID))}
				if err != nil {
					k8log.Error(err, "failed to retrieve json encoding of node")
				}
			}
			if keyLabel == "" {
				nnn[labelValue] = append(nnn[labelValue], jsn)
			} else if keyLabel == labelValue {
				nnn[labelValue] = append(nnn[labelValue], jsn)
			}
		} else if key.Key == "Deployment" {
			d := key.Value.(*appsv1.Deployment)
			labelValue := d.Labels[label]
			var jsn nodeMeta
			if meta {
				jsn = nodeMeta{Name: d.Name, Type: "workload", Id: base64.StdEncoding.EncodeToString([]byte(d.UID)), Value: d}
			} else {
				jsn = nodeMeta{Id: base64.StdEncoding.EncodeToString([]byte(d.UID))}
			}
			if keyLabel == "" {
				nnn[labelValue] = append(nnn[labelValue], jsn)
			} else if keyLabel == labelValue {
				nnn[labelValue] = append(nnn[labelValue], jsn)
			}
		}
	}

	return nnn
}

func (d data) getAnnotationData(annotation string) map[string][]string {
	nodes := make(map[string][]string)
	for key, _ := range d.nodes {
		if key.Key == "DeploymentConfig" {
			dc := key.Value.(*deploymentconfigv1.DeploymentConfig)
			var keys []string
			err := json.Unmarshal([]byte(dc.Annotations[annotation]), &keys)
			if err != nil {
				k8log.Error(err, "failed to retrieve json dencoding of node")
			}
			for _, key := range keys {
				json, err := json.Marshal(dc.UID)
				if err != nil {
					k8log.Error(err, "failed to retrieve json dencoding of node")
				}
				nodes[key] = append(nodes[key], string(json))
			}
		} else if key.Key == "Deployment" {
			d := key.Value.(*appsv1.Deployment)
			var keys []string
			err := json.Unmarshal([]byte(d.Annotations[annotation]), &keys)
			if err != nil {
				k8log.Error(err, "failed to retrieve json dencoding of node")
			}
			for _, key := range keys {
				jsn, err := json.Marshal(d.UID)
				if err != nil {
					k8log.Error(err, "failed to retrieve json dencoding of node")
				}
				nodes[key] = append(nodes[key], string(jsn))
			}
		}
	}

	return nodes
}

func formatNode(object interface{}, nodeType string) dataTypes {
	if nodeType == "DeploymentConfig" {
		dc := object.(*deploymentconfigv1.DeploymentConfig)
		return dataTypes{Id: base64.StdEncoding.EncodeToString([]byte(dc.UID)), Key: "DeploymentConfig", Value: object}
	}

	d := object.(*appsv1.Deployment)
	return dataTypes{Id: base64.StdEncoding.EncodeToString([]byte(d.UID)), Key: "Deployment", Value: object}
}

func getNode(event watch.Event) dataTypes {
	var x interface{} = event.Object
	var node dataTypes
	switch x.(type) {
	case *deploymentconfigv1.DeploymentConfig:
		node = formatNode(x, "DeploymentConfig")
	case *appsv1.Deployment:
		node = formatNode(x, "Deployment")
	default:
		fmt.Println(reflect.TypeOf(x))
	}

	return node
}

func getResource(rx interface{}) topology.Resource {
	var r topology.Resource
	switch rx.(type) {
	case *deploymentconfigv1.DeploymentConfig:
		r = formatDeploymentConfig(rx)
	case *appsv1.Deployment:
		r = formatDeployment(rx)
	case *corev1.Service:
		r = formatService(rx)
	case *routev1.Route:
		r = formatRoute(rx)
	case *corev1.ReplicationController:
		r = formatReplicationController(rx)
	case *appsv1.ReplicaSet:
		r = formatReplicaSet(rx)
	default:
		fmt.Println(reflect.TypeOf(rx))
	}

	return r
}

func addOrUpdate(slice []topology.Resource, i topology.Resource) []topology.Resource {
	for index, ele := range slice {
		if ele == i {
			slice[index] = i
			return slice
		}
	}
	return append(slice, i)
}
