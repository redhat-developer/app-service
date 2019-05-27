package appserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	deploymentconfigv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	deploymentconfigclientset "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"github.com/redhat-developer/app-service/appserver/topology"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var k8log = logf.Log

type resource struct {
	Metadata string `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Status   string `json:"status,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Name     string `json:"name,name=name"`
	Kind     string `json:"kind,name=kind"`
}

type node struct {
	Id   string
	Name string
}
type nodeData struct {
	NodeType  string     `json:"nodeType,name=nodeType"`
	Id        string     `json:"id,name=id"`
	Resources []resource `json:"resource,omitempty" protobuf:"bytes,1,opt,name=resource"`
	Data      staticData
}

type group struct {
	Id    string
	Name  string
	Nodes []string
}

type edge struct {
	Source string
	Target string
}

type nodeMeta struct {
	Id   string
	Name string
	Type string
}

type staticData struct {
	url          string
	editUrl      string
	builderImage string
	donutStatus  string
}

type data struct {
	nodes []interface{}
}

// HandleTopology returns the handler function for the /status endpoint
func (srv *AppServer) HandleTopology() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		namespace := "default"
		host := "https://api.tkurian11.devcluster.openshift.com:6443"
		bearerToken := "axSscPvCfX-j6F2xi1twnfp15WmMMp8RmbOow66v89Q"

		openshiftAPIConfig := getOpenshiftAPIConfig(host, bearerToken)

		cl, _ := kubernetes.NewForConfig(&openshiftAPIConfig)
		clDepConfig, _ := deploymentconfigclientset.NewForConfig(&openshiftAPIConfig)
		clRoute, _ := routeclientset.NewForConfig(&openshiftAPIConfig)
		dc := getNodes(clDepConfig, cl, namespace)

		nodes := dc.getUniqueNodes()
		groups := dc.getGroups()
		edges := dc.getEdges()
		formattedDc := dc.formatNodes()

		resources := getResources(nodes, cl, clRoute, clDepConfig, namespace)
		result := topology.GetSampleTopology(formattedDc, resources, groups, edges)

		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
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

func getNodes(clDepConfig *deploymentconfigclientset.AppsV1Client, cl *kubernetes.Clientset, namespace string) data {
	// Store all node types
	var n []interface{}

	deploymentConfigs, err := clDepConfig.DeploymentConfigs(namespace).List(metav1.ListOptions{})
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of node")
	}
	n = append(n, deploymentConfigs.Items)

	deployments, err := cl.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		k8log.Error(err, "failed to retrieve json encoding of node")
	}
	n = append(n, deployments.Items)

	return data{nodes: n}
}

func (d data) getUniqueNodes() map[string][]string {
	return d.getLabelData("app.kubernetes.io/name", "", false)
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

				targetId, err := json.Marshal(target)
				if err != nil {
					k8log.Error(err, "failed to retrieve json encoding of node")
				}

				e, err := json.Marshal(edge{Source: source, Target: string(targetId)})
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

	nodes = d.getLabelData("app.kubernetes.io/part-of", "", true)

	for key, value := range nodes {
		g, err := json.Marshal(group{Id: "1", Name: key, Nodes: value})
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of node")
		}
		groups = append(groups, string(g))
	}

	return groups
}

func getResources(dc map[string][]string, cl *kubernetes.Clientset, clRoute *routeclientset.RouteV1Client, clDepConfig *deploymentconfigclientset.AppsV1Client, namespace string) map[string]string {
	nodeDatas := make(map[string]string)

	for labelKey, dcNodes := range dc {
		options := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", labelKey),
		}
		fmt.Println(labelKey)
		for _, dc := range dcNodes {
			var nm nodeMeta
			err := json.Unmarshal([]byte(dc), &nm)
			if err != nil {
				k8log.Error(err, "failed to list existing replicationControllers")
			}

			//Replication Controller
			replicationControllers, err := cl.CoreV1().ReplicationControllers(namespace).List(options)
			if err != nil {
				k8log.Error(err, "failed to list existing deployment configs")
			}
			resources := formatReplicationControllers(replicationControllers.Items)
			fmt.Println(replicationControllers.Items)
			//Services
			services, err := cl.CoreV1().Services(namespace).List(options)
			if err != nil {
				k8log.Error(err, "failed to list existing deployment configs")
			}
			resources = append(resources, formatService(services.Items)...)

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

			//fmt.Println(resources)
			nd, err := json.Marshal(nodeData{Id: nm.Id, NodeType: nm.Type, Resources: resources, Data: staticData{url: "dummy_url", editUrl: "dummy_edit_url", builderImage: "deploymentsLabels['app.kubernetes.io/name']", donutStatus: ""}})
			if err != nil {
				k8log.Error(err, "failed to retrieve json encoding of node")
			}
			nodeDatas[nm.Id] = string(nd)
		}
	}
	return nodeDatas
}

func formatDeploymentConfigs(deploymentConfigItems []deploymentconfigv1.DeploymentConfig) []resource {
	var resources []resource

	for _, elem := range deploymentConfigItems {
		meta, err := json.Marshal(elem.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of deployment config")
		}
		status, err := json.Marshal(elem.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of deployment config")
		}
		resources = append(resources, resource{Name: elem.Name, Kind: elem.Kind, Metadata: string(meta), Status: string(status)})
	}
	return resources
}

func formatService(services []corev1.Service) []resource {
	var resources []resource

	for _, elem := range services {
		meta, err := json.Marshal(elem.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of service")
		}
		status, err := json.Marshal(elem.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of service")
		}
		resources = append(resources, resource{Name: elem.Name, Kind: elem.Kind, Metadata: string(meta), Status: string(status)})
	}
	return resources
}

func formatReplicationControllers(replicaSetItems []corev1.ReplicationController) []resource {
	var resources []resource

	for _, elem := range replicaSetItems {
		meta, err := json.Marshal(elem.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of replication controller")
		}
		status, err := json.Marshal(elem.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of replication controller")
		}
		resources = append(resources, resource{Name: elem.Name, Kind: elem.Kind, Metadata: string(meta), Status: string(status)})
	}
	return resources
}

func formatRoutes(routeItems []routev1.Route) []resource {
	var resources []resource

	for _, elem := range routeItems {
		meta, err := json.Marshal(elem.GetObjectMeta())
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of route")
		}
		status, err := json.Marshal(elem.Status)
		if err != nil {
			k8log.Error(err, "failed to retrieve json encoding of route")
		}
		resources = append(resources, resource{Name: elem.Name, Kind: elem.Kind, Metadata: string(meta), Status: string(status)})
	}
	return resources
}

func (d data) formatNodes() []string {
	var nodes []string

	for _, elem := range d.nodes {
		switch elem.(type) {
		case []deploymentconfigv1.DeploymentConfig:
			deploymentsConfigs := elem.([]deploymentconfigv1.DeploymentConfig)
			for _, dc := range deploymentsConfigs {
				n, err := json.Marshal(node{Name: dc.Name, Id: base64.StdEncoding.EncodeToString([]byte(dc.UID))})
				if err != nil {
					k8log.Error(err, "failed to retrieve json encoding of node")
				}
				nodes = append(nodes, string(n))
			}
		case []appsv1.Deployment:
			deployments := elem.([]appsv1.Deployment)
			for _, d := range deployments {
				n, err := json.Marshal(node{Name: d.Name, Id: base64.StdEncoding.EncodeToString([]byte(d.UID))})
				if err != nil {
					k8log.Error(err, "failed to retrieve json encoding of node")
				}
				nodes = append(nodes, string(n))
			}
		}

	}
	return nodes
}

func (d data) getLabelData(label string, key string, ok bool) map[string][]string {

	nodes := make(map[string][]string)
	for _, elem := range d.nodes {
		switch elem.(type) {
		case []deploymentconfigv1.DeploymentConfig:
			deploymentsConfigs := elem.([]deploymentconfigv1.DeploymentConfig)
			for _, dc := range deploymentsConfigs {
				labelValue := dc.Labels[label]
				var jsn []byte
				var err error
				if ok {
					jsn, err = json.Marshal(dc.UID)
					if err != nil {
						k8log.Error(err, "failed to retrieve json encoding of node")
					}
				} else {
					jsn, err = json.Marshal(nodeMeta{Name: dc.Name, Type: "workload", Id: base64.StdEncoding.EncodeToString([]byte(dc.UID))})
					if err != nil {
						k8log.Error(err, "failed to retrieve json encoding of node")
					}
				}
				if key == "" {
					nodes[labelValue] = append(nodes[labelValue], string(jsn))
				} else if key == labelValue {
					nodes[labelValue] = append(nodes[labelValue], string(jsn))
				}
			}
		case []appsv1.Deployment:
			deployments := elem.([]appsv1.Deployment)
			for _, d := range deployments {
				labelValue := d.Labels[label]
				jsn, err := json.Marshal(d.UID)
				if err != nil {
					k8log.Error(err, "failed to retrieve json encoding of node")
				}

				if key == "" {
					nodes[labelValue] = append(nodes[labelValue], string(jsn))
				} else if key == labelValue {
					nodes[labelValue] = append(nodes[labelValue], string(jsn))
				}
			}
		}
	}

	return nodes
}

func (d data) getAnnotationData(annotation string) map[string][]string {
	nodes := make(map[string][]string)
	for _, elem := range d.nodes {
		switch elem.(type) {
		case []deploymentconfigv1.DeploymentConfig:
			deploymentsConfigs := elem.([]deploymentconfigv1.DeploymentConfig)
			for _, dc := range deploymentsConfigs {
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
			}
		case []appsv1.Deployment:
			deployments := elem.([]appsv1.Deployment)
			for _, d := range deployments {
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
	}

	return nodes
}
