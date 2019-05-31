package kubeclient

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (kc KubeClient) GetDeploymentWatcher(namespace string, options v1.ListOptions, onError func(err error)) watch.Interface {
	w, err := kc.CoreClient.AppsV1().Deployments(namespace).Watch(options)
	if err != nil {
		onError(err)
	}
	return w
}

func (kc KubeClient) GetDeploymentConfigWatcher(namespace string, options v1.ListOptions, onError func(err error)) watch.Interface {
	w, err := kc.OcAppsClient.DeploymentConfigs(namespace).Watch(options)
	if err != nil {
		onError(err)
	}
	return w
}

func (kc KubeClient) GetReplicationControllerWatcher(namespace string, options v1.ListOptions, onError func(err error)) watch.Interface {
	w, err := kc.CoreClient.CoreV1().ReplicationControllers(namespace).Watch(options)
	if err != nil {
		onError(err)
	}
	return w
}

func (kc KubeClient) GetPodWatcher(namespace string, options v1.ListOptions, onError func(err error)) watch.Interface {
	w, err := kc.CoreClient.CoreV1().Pods(namespace).Watch(options)
	if err != nil {
		onError(err)
	}
	return w
}

func (kc KubeClient) GetRouteWatcher(namespace string, options v1.ListOptions, onError func(err error)) watch.Interface {
	w, err := kc.OcRouteClient.Routes(namespace).Watch(options)
	if err != nil {
		onError(err)
	}
	return w
}

func (kc KubeClient) GetServiceWatcher(namespace string, options v1.ListOptions, onError func(err error)) watch.Interface {
	w, err := kc.CoreClient.CoreV1().Services(namespace).Watch(options)
	if err != nil {
		onError(err)
	}
	return w
}
