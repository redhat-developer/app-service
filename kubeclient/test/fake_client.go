package test

import (
	ocfakeappsclient "github.com/openshift/client-go/apps/clientset/versioned/fake"
	ocfakerouteclient "github.com/openshift/client-go/route/clientset/versioned/fake"
	"github.com/redhat-developer/app-service/kubeclient"
	"k8s.io/client-go/kubernetes/fake"
)

func FakeKubeClient() *kubeclient.KubeClient {
	k := &kubeclient.KubeClient{}
	k.CoreClient = fake.NewSimpleClientset()
	k.OcRouteClient = ocfakerouteclient.NewSimpleClientset().RouteV1()
	k.OcAppsClient = ocfakeappsclient.NewSimpleClientset().AppsV1()
	return k
}
