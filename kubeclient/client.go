package kubeclient

import (
	"flag"
	ocappsclient "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	ocrouteclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type KubeClient struct {
   	CoreClient    kubernetes.Interface
   	OcAppsClient  ocappsclient.AppsV1Interface
	OcRouteClient ocrouteclient.RouteV1Interface
}

func NewKubeClient(config *rest.Config) *KubeClient  {
	var err error
	kc := new(KubeClient)
	kc.CoreClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	kc.OcAppsClient, err = ocappsclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	kc.OcRouteClient, err = ocrouteclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return kc;
}

func GetKubeConfig() *rest.Config {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	return config
}