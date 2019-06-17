package appserver

import (
	"fmt"
	"testing"

	"github.com/redhat-developer/app-service/appserver/topology"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppServer_GetResources(t *testing.T) {
	var nMap nodesMap
	var iData innerData
	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = "testapp"

	iData.nd.ID = "1"
	iData.nd.Type = "workload"
	iData.nd.Data.BuilderImage = "test"
	iData.nd.Data.DonutStatus = make(map[string]string)
	iData.nd.Data.EditURL = "https://test/url"
	iData.nd.Data.URL = "https://test/url"
	iData.nd.Name = "testapp"
	iData.nm.ID = "1"
	iData.nm.Labels = labels
	iData.nm.Name = "testapp"
	iData.nm.Type = "workload"
	iData.nm.Kind = "Service"
	iData.nm.Value = make(map[string]string)

	expected := make(map[string]string)
	expected["1"] = "{\"name\":\"testapp\",\"type\":\"workload\",\"id\":\"1\",\"data\":{\"url\":\"https://test/url\",\"editUrl\":\"https://test/url\",\"builderImage\":\"test\",\"donutStatus\":{}}}"

	nMap.nodes = make(map[string]innerData)
	nMap.nodes["testing"] = iData

	resources := nMap.getResources()
	require.Equal(t, expected, resources)
}

func TestAppServer_GetEdges(t *testing.T) {
	var nMap nodesMap
	nginx := createResource("1", "DeploymentConfig", "nginx", "testapp", "nodejs")
	nodejs := createResource("2", "DeploymentConfig", "nodejs", "testapp", "")

	t.Log(nginx.nm.Annotations)

	nMap.nodes = make(map[string]innerData)
	nMap.nodes["nginx"] = nginx
	nMap.nodes["nodejs"] = nodejs

	edges := nMap.getEdges()
	require.Equal(t, "2", edges[0].ID)
	require.Equal(t, "2", edges[0].Source)
	require.Equal(t, "1", edges[0].Target)
	require.Equal(t, "connects-to", edges[0].Type)

}

func TestAppServer_GetGroups(t *testing.T) {
	var nMap nodesMap
	nginx := createResource("1", "DeploymentConfig", "nginx", "testapp", "nodejs")
	nMap.nodes = make(map[string]innerData)
	nMap.nodes["nginx"] = nginx

	groups := nMap.getGroups()

	require.Equal(t, "testapp", groups[0].Name)
	require.Equal(t, "group:testapp", groups[0].ID)
}

func TestAppServer_GetNode(t *testing.T) {
	var nMap nodesMap
	nginx := createResource("1", "DeploymentConfig", "nginx", "testapp", "nodejs")
	nodejs := createResource("2", "DeploymentConfig", "nodejs", "testapp", "")

	nMap.nodes = make(map[string]innerData)
	nMap.nodes["nginx"] = nginx
	nMap.nodes["nodejs"] = nodejs

	nodes := nMap.getNode()
	var expected []string
	expected = append(expected, "{\"id\":\"1\",\"name\":\"nginx\"}", "{\"id\":\"2\",\"name\":\"nodejs\"}")

	require.Equal(t, expected, nodes)
}

func TestAppServer_GetLabelData(t *testing.T) {
	var nMap nodesMap
	nginx := createResource("1", "DeploymentConfig", "nginx", "testapp", "nodejs")
	nodejs := createResource("2", "DeploymentConfig", "nodejs", "testapp", "")

	nMap.nodes = make(map[string]innerData)
	nMap.nodes["nginx"] = nginx
	nMap.nodes["nodejs"] = nodejs

	labelData := nMap.getLabelData("app.kubernetes.io/name", "")

	expected := make(map[string][]nodeMeta)
	expected["nginx"] = append(expected["nginx"], nginx.nm)
	expected["nodejs"] = append(expected["nodejs"], nodejs.nm)

	require.Equal(t, expected, labelData)
}

func TestAppServer_GetAnnotationData(t *testing.T) {
	var nMap nodesMap
	nginx := createResource("1", "DeploymentConfig", "nginx", "testapp", "nodejs")
	nodejs := createResource("2", "DeploymentConfig", "nodejs", "testapp", "")

	nMap.nodes = make(map[string]innerData)
	nMap.nodes["nginx"] = nginx
	nMap.nodes["nodejs"] = nodejs

	annotationData := nMap.getAnnotationData("app.openshift.io/connects-to")

	expected := make(map[string][]string)
	expected["nodejs"] = append(expected["nodejs"], "1")

	require.Equal(t, expected, annotationData)
}

func TestAppServer_DeleteNode(t *testing.T) {
	var nMap nodesMap
	nginx := createResource("1", "DeploymentConfig", "nginx", "testapp", "nodejs")
	nodejs := createResource("2", "DeploymentConfig", "nodejs", "testapp", "")

	nMap.nodes = make(map[string]innerData)
	nMap.nodes["nginx"] = nginx
	nMap.nodes["nodejs"] = nodejs

	nMap.deleteNode(nodejs.nm)

	expected := make(map[string][]string)
	expected["nodejs"] = append(expected["nodejs"], "1")

	require.Equal(t, "", nMap.nodes["nodejs"].nm.ID)
	require.Equal(t, "1", nMap.nodes["nginx"].nm.ID)
}

func TestAppServer_AddOrUpdateNode(t *testing.T) {
	var nMap nodesMap
	nginx := createResource("1", "DeploymentConfig", "nginx", "testapp", "nodejs")
	nodejs := createResource("2", "DeploymentConfig", "nodejs", "testapp", "")

	newNodejs := createResource("3", "DeploymentConfig", "nodejs", "testapp", "")
	perl := createResource("4", "DeploymentConfig", "perl", "testapp", "")

	nMap.nodes = make(map[string]innerData)
	nMap.nodes["nginx"] = nginx
	nMap.nodes["nodejs"] = nodejs

	// Test updating nodejs
	nMap.addOrUpdateNode(newNodejs.nm)

	require.Equal(t, "3", nMap.nodes["nodejs"].nm.ID)
	require.Equal(t, "1", nMap.nodes["nginx"].nm.ID)
	require.Equal(t, "", nMap.nodes["perl"].nm.ID)

	// Test adding perl
	nMap.addOrUpdateNode(perl.nm)

	require.Equal(t, "3", nMap.nodes["nodejs"].nm.ID)
	require.Equal(t, "1", nMap.nodes["nginx"].nm.ID)
	require.Equal(t, "4", nMap.nodes["perl"].nm.ID)
}

func TestAppServer_DeleteNodeResource(t *testing.T) {
	var nMap nodesMap
	nodejs := createResource("2", "DeploymentConfig", "nodejs", "testapp", "")

	var resourceService topology.Resource
	var resourceDeploymentConfig topology.Resource

	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = "testapp"

	resourceService.Kind = "Service"
	resourceService.Metadata = "{}"
	resourceService.Name = "nodejs"
	resourceService.Status = "{}"

	resourceDeploymentConfig.Kind = "DeploymentConfig"
	resourceDeploymentConfig.Metadata = "{}"
	resourceDeploymentConfig.Name = "nodejs"
	resourceDeploymentConfig.Status = "{}"

	nMap.nodes = make(map[string]innerData)
	nMap.nodes["nodejs"] = nodejs
	nMap.addOrUpdateNodeResource("nodejs", resourceDeploymentConfig)
	nMap.addOrUpdateNodeResource("nodejs", resourceService)

	require.Equal(t, 2, len(nMap.nodes["nodejs"].nd.Resources))

	nMap.deleteNodeResource(nodejs.nm, resourceService)

	require.Equal(t, 1, len(nMap.nodes["nodejs"].nd.Resources))
	require.Equal(t, "nodejs", nMap.nodes["nodejs"].nd.Resources[0].Name)
	require.Equal(t, "DeploymentConfig", nMap.nodes["nodejs"].nd.Resources[0].Kind)
}

func TestAppServer_AddOrUpdateNodeResource(t *testing.T) {
	var nMap nodesMap
	nodejs := createResource("2", "DeploymentConfig", "nodejs", "testapp", "")
	var resourceDeploymentConfig topology.Resource
	var newDeploymentConfig topology.Resource

	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = "testapp"

	resourceDeploymentConfig.Kind = "DeploymentConfig"
	resourceDeploymentConfig.Metadata = "{}"
	resourceDeploymentConfig.Name = "nodejs"
	resourceDeploymentConfig.Status = "{}"

	newDeploymentConfig.Kind = "DeploymentConfig"
	newDeploymentConfig.Metadata = "{\"test\": \"test\"}"
	newDeploymentConfig.Name = "nodejs"
	newDeploymentConfig.Status = "{}"

	nMap.nodes = make(map[string]innerData)
	nMap.nodes["nodejs"] = nodejs
	nMap.addOrUpdateNodeResource("nodejs", resourceDeploymentConfig)

	require.Equal(t, 1, len(nMap.nodes["nodejs"].nd.Resources))
	require.Equal(t, "nodejs", nMap.nodes["nodejs"].nd.Resources[0].Name)
	require.Equal(t, "DeploymentConfig", nMap.nodes["nodejs"].nd.Resources[0].Kind)
	require.Equal(t, "{}", nMap.nodes["nodejs"].nd.Resources[0].Metadata)

	nMap.addOrUpdateNodeResource("nodejs", newDeploymentConfig)

	require.Equal(t, 1, len(nMap.nodes["nodejs"].nd.Resources))
	require.Equal(t, "nodejs", nMap.nodes["nodejs"].nd.Resources[0].Name)
	require.Equal(t, "DeploymentConfig", nMap.nodes["nodejs"].nd.Resources[0].Kind)
	require.Equal(t, "{\"test\": \"test\"}", nMap.nodes["nodejs"].nd.Resources[0].Metadata)
}

func TestAppServer_GetResourcesListOptions(t *testing.T) {
	nMeta := make(map[string][]nodeMeta)
	nginx := createResource("1", "DeploymentConfig", "nginx", "testapp", "nodejs")
	nodejs := createResource("2", "DeploymentConfig", "nodejs", "testapp", "")

	nMeta["nginx"] = append(nMeta["nginx"], nginx.nm)
	nMeta["nodejs"] = append(nMeta["nodejs"], nodejs.nm)

	listOptions := getResourcesListOptions(nMeta)
	optionsNodeJS := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", "nodejs"),
	}

	require.Equal(t, "nodejs", listOptions[optionsNodeJS].Name)
}

func createResource(id string, kind string, name string, partOf string, annotation string) innerData {
	var iData innerData

	labels := make(map[string]string)
	annotations := make(map[string]string)
	labels["app.kubernetes.io/name"] = name
	labels["app.kubernetes.io/part-of"] = partOf

	iData.nd.ID = id
	iData.nd.Type = "workload"
	iData.nd.Data.BuilderImage = "test"
	iData.nd.Data.DonutStatus = make(map[string]string)
	iData.nd.Data.EditURL = "https://test/url"
	iData.nd.Data.URL = "https://test/url"
	iData.nm.ID = id
	iData.nm.Labels = labels
	iData.nm.Name = name
	iData.nm.Type = "workload"
	iData.nm.Kind = kind
	iData.nm.Value = make(map[string]string)
	if annotation != "" {
		annotations["app.openshift.io/connects-to"] = "[" + "\"" + annotation + "\"" + "]"
		iData.nm.Annotations = annotations
	}

	return iData
}
