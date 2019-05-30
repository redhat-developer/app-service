package kubeclient

import (
	ocfakeappsclient "github.com/openshift/client-go/apps/clientset/versioned/fake"
	ocfakerouteclient "github.com/openshift/client-go/route/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestNewKubeClient(t *testing.T) {
	k := KubeClient{}
	k.CoreClient = fake.NewSimpleClientset()
	k.OcRouteClient = ocfakerouteclient.NewSimpleClientset().RouteV1()
	k.OcAppsClient = ocfakeappsclient.NewSimpleClientset().AppsV1()
	assert.NotNil(t, k.CoreClient, "Kubecore client shouldn't be nil")
}

