package appserver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	deploymentconfigv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	ocappsclient "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"github.com/redhat-developer/app-service/appserver/topology"
	"github.com/redhat-developer/app-service/kubeclient"
	"github.com/redhat-developer/app-service/watcher"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var k8log = logf.Log

type myNodes struct {
	nodes map[string]map[string][]string
}

type nodeMeta struct {
	Id   string
	Name string
	Type string
}

type dataTypes struct {
	Id    string
	Key   string
	Value interface{}
}

type data struct {
	nodes map[dataTypes][]dataTypes
}

// HandleTopology returns the handler function for the /status endpoint
func (srv *AppServer) HandleTopology() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		namespace := "default"
		host := "https://api.tkurian16.devcluster.openshift.com:6443"
		bearerToken := "7DdF7VrdYl2F9MrR95J_v0Z0pJj1qh6tMZrSzbn_Uno"

		openshiftAPIConfig := getOpenshiftAPIConfig(host, bearerToken)

		k := kubeclient.NewKubeClient()
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

		var result topology.VisualizationResponse
		var d data
		newWatch.StartWatcher()
		testMap := make(map[dataTypes][]dataTypes)

		newWatch.ListenWatcher(func(event watch.Event) {

			var x interface{} = event.Object
			var node dataTypes
			switch x.(type) {
			case *deploymentconfigv1.DeploymentConfig:
				node = createNode(x, "DeploymentConfig")
			case *appsv1.Deployment:
				node = createNode(x, "Deployment")
			default:
				fmt.Println(reflect.TypeOf(x))
			}

			testMap[node] = append(testMap[node], dataTypes{})
			d.nodes = testMap

			nodes := d.getUniqueNodes()
			groups := d.getGroups()
			edges := d.getEdges()
			formattedDc := d.formatNodes()

			clDepConfig, _ := ocappsclient.NewForConfig(&openshiftAPIConfig)
			clRoute, _ := routeclientset.NewForConfig(&openshiftAPIConfig)
			getResources(nodes, k.CoreClient, clRoute, clDepConfig, namespace)
			resourceOptsList := getResourcesListOptions(nodes)
			resourceWatchers := make(map[nodeMeta]*watcher.Watch)
			nodeDatas := make(map[string]string)
			nodeDataNotMarshalled := make(map[string]topology.NodeData)
			//--------------------------------------------------------------
			//--------------------------------------------------------------
			for opts, v := range resourceOptsList {
				newResourceWatch := watcher.NewWatch(namespace,
					k,
					k.GetDeploymentConfigWatcher(namespace, opts, onGetWatchError),
					k.GetReplicationControllerWatcher(namespace, opts, onGetWatchError),
					k.GetServiceWatcher(namespace, opts, onGetWatchError),
					k.GetRouteWatcher(namespace, opts, onGetWatchError),
				)
				newResourceWatch.SetFilters([]watch.EventType{watch.Added, watch.Modified})

				resourceWatchers[v] = newResourceWatch
			}

			//nodeResources := make(map[nodeMeta][]topology.Resource)
			for nm, v := range resourceWatchers {
				v.StartWatcher()
				v.ListenWatcher(func(resourceEvent watch.Event) {
					var rx interface{} = resourceEvent.Object
					fmt.Println(rx)
					fmt.Println(nm.Id)

					// figure out type
					// call format for type
					var r topology.Resource
					switch x.(type) {
					case *deploymentconfigv1.DeploymentConfig:
						r = formatDeploymentConfig(x)
					case *appsv1.Deployment:
						//node = createResource(x, "Deployment")
					case *corev1.Service:
						r = formatService(x)
					case *routev1.Route:
						r = formatRoute(x)
					case *corev1.ReplicationController:
						r = formatReplicationController(x)
					default:
						fmt.Println(reflect.TypeOf(x))
					}

					item := nodeDataNotMarshalled[nm.Id]

					if item.Id == "" {
						item.Resources = append(item.Resources, r)

						nd, err := json.Marshal(item)
						if err != nil {
							k8log.Error(err, "failed to retrieve json encoding of node")
						}
						nodeDatas[nm.Id] = string(nd)
						// for this node, find the corresponding: nd, err := json.Marshal(topology.NodeData{Id: nm.Id, Type: nm.Type, Resources: resources, Data: topology.Data{Url: "dummy_url", EditUrl: "dummy_edit_url", BuilderImage: labelKey, DonutStatus: make(map[string]string)}})
						// the nd's need to be stored in a map for easy look up and then we are going to modify the resource section by adding the resource.
					} else {
						var res []topology.Resource
						res = append(res, r)
						topResource := topology.NodeData{Id: nm.Id, Type: nm.Type, Resources: res, Data: topology.Data{Url: "dummy_url", EditUrl: "dummy_edit_url", BuilderImage: "TODO: TINA FIX", DonutStatus: make(map[string]string)}}
						nodeDataNotMarshalled[nm.Id] = topResource
						nd, err := json.Marshal(topResource)
						if err != nil {
							k8log.Error(err, "failed to retrieve json encoding of node")
						}
						nodeDatas[nm.Id] = string(nd)
					}

					result = topology.GetSampleTopology(formattedDc, nodeDatas, groups, edges)
					by, _ := json.Marshal(result)
					b := bytes.NewBuffer(by)

					w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/json")
					if _, err := b.WriteTo(w); err != nil {
						fmt.Fprintf(w, "%s", err)
					}

					w.Write(by)
				})
			}

			//--------------------------------------------------------------
			//--------------------------------------------------------------

		})
	}
}

func getOpenshiftAPIConfig(host string, bearerToken string) rest.Config {
	return rest.Config{
		Host:        host,
		BearerToken: bearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
}

func (d data) getUniqueNodes() map[string][]string {
	return d.getLabelData("app.kubernetes.io/name", "", true)
}

func (d data) getEdges() []string {
	var edges []string
	sourceObjects := make(map[string][]string)
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
				var nm nodeMeta
				err := json.Unmarshal([]byte(source), &nm)
				if err != nil {
					k8log.Error(err, "failed to get node data")
				}

				e, err := json.Marshal(topology.Edge{Source: nm.Id, Target: target})
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
	nodes := make(map[string][]string)
	var groups []string
	var gs []string

	nodes = d.getLabelData("app.kubernetes.io/part-of", "", false)
	for key, value := range nodes {
		for _, v := range value {
			gs = append(gs, string(v))
		}
		g, err := json.Marshal(topology.Group{Id: "group:" + key, Name: key, Nodes: gs})
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of node")
		}
		groups = append(groups, string(g))
	}

	return groups
}

func getResourcesListOptions(dc map[string][]string) map[metav1.ListOptions]nodeMeta {
	listOptions := make(map[metav1.ListOptions]nodeMeta)
	for labelKey, dcNodes := range dc {
		options := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", labelKey),
		}
		for _, dc := range dcNodes {
			var nm nodeMeta
			err := json.Unmarshal([]byte(dc), &nm)
			if err != nil {
				k8log.Error(err, "failed to list existing replicationControllers")
			}
			listOptions[options] = nm
		}

	}
	return listOptions
}

func getResources(dc map[string][]string, cl kubernetes.Interface, clRoute *routeclientset.RouteV1Client, clDepConfig *ocappsclient.AppsV1Client, namespace string) map[string]string {
	nodeDatas := make(map[string]string)
	for labelKey, dcNodes := range dc {
		options := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", labelKey),
		}
		for _, dc := range dcNodes {
			var nm nodeMeta
			err := json.Unmarshal([]byte(dc), &nm)
			if err != nil {
				k8log.Error(err, "failed to list existing replicationControllers")
			}

			//Replication Controllers
			replicationControllers, err := cl.CoreV1().ReplicationControllers(namespace).List(options)
			if err != nil {
				k8log.Error(err, "failed to list existing deployment configs")
			}
			resources := formatReplicationControllers(replicationControllers.Items)
			//Services
			services, err := cl.CoreV1().Services(namespace).List(options)
			if err != nil {
				k8log.Error(err, "failed to list existing deployment configs")
			}
			resources = append(resources, formatServices(services.Items)...)

			// Routes
			routes, err := clRoute.Routes(namespace).List(options)
			if err != nil {
				k8log.Error(err, "failed to list existing routes")
			}
			resources = append(resources, formatRoutes(routes.Items)...)

			// DeploymentConfigs
			deploymentConfigs, err := clDepConfig.DeploymentConfigs(namespace).List(options)
			if err != nil {
				k8log.Error(err, "failed to list existing deployment configs")
			}
			resources = append(resources, formatDeploymentConfigs(deploymentConfigs.Items)...)

			nd, err := json.Marshal(topology.NodeData{Id: nm.Id, Type: nm.Type, Resources: resources, Data: topology.Data{Url: "dummy_url", EditUrl: "dummy_edit_url", BuilderImage: labelKey, DonutStatus: make(map[string]string)}})
			if err != nil {
				k8log.Error(err, "failed to retrieve json encoding of node")
			}
			nodeDatas[nm.Id] = string(nd)
		}
	}
	return nodeDatas
}

func formatDeploymentConfigs(deploymentConfigItems []deploymentconfigv1.DeploymentConfig) []topology.Resource {
	var resources []topology.Resource

	for _, elem := range deploymentConfigItems {
		meta, err := json.Marshal(elem.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of deployment config")
		}
		status, err := json.Marshal(elem.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of deployment config")
		}
		resources = append(resources, topology.Resource{Name: elem.Name, Kind: elem.Kind, Metadata: string(meta), Status: string(status)})
	}
	return resources
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
	return topology.Resource{Name: deploymentConfig.Name, Kind: deploymentConfig.Kind, Metadata: string(meta), Status: string(status)}
}

func formatServices(services []corev1.Service) []topology.Resource {
	var resources []topology.Resource

	for _, elem := range services {
		meta, err := json.Marshal(elem.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of service")
		}
		status, err := json.Marshal(elem.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of service")
		}
		resources = append(resources, topology.Resource{Name: elem.Name, Kind: elem.Kind, Metadata: string(meta), Status: string(status)})
	}
	return resources
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
	return topology.Resource{Name: serv.Name, Kind: serv.Kind, Metadata: string(meta), Status: string(status)}
}

func formatReplicationControllers(replicaSetItems []corev1.ReplicationController) []topology.Resource {
	var resources []topology.Resource

	for _, elem := range replicaSetItems {
		meta, err := json.Marshal(elem.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of replication controller")
		}
		status, err := json.Marshal(elem.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of replication controller")
		}
		resources = append(resources, topology.Resource{Name: elem.Name, Kind: elem.Kind, Metadata: string(meta), Status: string(status)})
	}
	return resources
}

func formatReplicationController(replicaSetItems interface{}) topology.Resource {
	rc := replicaSetItems.(*corev1.ReplicationController)
	meta, err := json.Marshal(rc.GetObjectMeta())
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of replication controller")
	}
	status, err := json.Marshal(rc.Status)
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of replication controller")
	}
	return topology.Resource{Name: rc.Name, Kind: rc.Kind, Metadata: string(meta), Status: string(status)}
}

func formatRoutes(routeItems []routev1.Route) []topology.Resource {
	var resources []topology.Resource

	for _, elem := range routeItems {
		meta, err := json.Marshal(elem.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of route")
		}
		status, err := json.Marshal(elem.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of route")
		}
		resources = append(resources, topology.Resource{Name: elem.Name, Kind: elem.Kind, Metadata: string(meta), Status: string(status)})
	}
	return resources
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
	return topology.Resource{Name: r.Name, Kind: r.Kind, Metadata: string(meta), Status: string(status)}
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

func (d data) getLabelData(label string, keyLabel string, meta bool) map[string][]string {
	nnn := make(map[string][]string)
	for key, _ := range d.nodes {
		if key.Key == "DeploymentConfig" {
			dc := key.Value.(*deploymentconfigv1.DeploymentConfig)
			labelValue := dc.Labels[label]
			var jsn []byte
			var err error
			if meta {
				jsn, err = json.Marshal(nodeMeta{Name: dc.Name, Type: "workload", Id: base64.StdEncoding.EncodeToString([]byte(dc.UID))})
				if err != nil {
					k8log.Error(err, "failed to retrieve json encoding of node")
				}
			} else {
				jsn, err = json.Marshal(dc.UID)
				if err != nil {
					k8log.Error(err, "failed to retrieve json encoding of node")
				}
			}
			if keyLabel == "" {
				nnn[labelValue] = append(nnn[labelValue], string(jsn))
			} else if keyLabel == labelValue {
				nnn[labelValue] = append(nnn[labelValue], string(jsn))
			}
		} else if key.Key == "Deployment" {
			d := key.Value.(*appsv1.Deployment)
			labelValue := d.Labels[label]
			jsn, err := json.Marshal(d.UID)
			if err != nil {
				k8log.Error(err, "failed to retrieve json encoding of node")
			}

			if keyLabel == "" {
				nnn[labelValue] = append(nnn[labelValue], string(jsn))
			} else if keyLabel == labelValue {
				nnn[labelValue] = append(nnn[labelValue], string(jsn))
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

func createNode(object interface{}, nodeType string) dataTypes {
	if nodeType == "DeploymentConfig" {
		dc := object.(*deploymentconfigv1.DeploymentConfig)
		return dataTypes{Id: base64.StdEncoding.EncodeToString([]byte(dc.UID)), Key: "DeploymentConfig", Value: object}
	}

	d := object.(*appsv1.Deployment)
	return dataTypes{Id: base64.StdEncoding.EncodeToString([]byte(d.UID)), Key: "Deployment", Value: object}
}
