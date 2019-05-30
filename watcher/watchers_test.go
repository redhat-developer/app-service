package watcher

import (
	"fmt"
	"github.com/redhat-developer/app-service/kubeclient/test"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"testing"
)

func int32Ptr(i int32) *int32 { return &i }

func TestNewWatch(t *testing.T) {
	k := test.FakeKubeClient()
	namespace := "myproject"
	listOptions := metav1.ListOptions{}
	onGetWatchError := func(err error) {
		fmt.Errorf("Error is %+v", err)
	}
	newWatch := NewWatch(namespace,
		k,
		k.GetDeploymentConfigWatcher(namespace, listOptions, onGetWatchError),
		k.GetPodWatcher(namespace, listOptions, onGetWatchError),
		k.GetRouteWatcher(namespace, listOptions, onGetWatchError),
		k.GetDeploymentWatcher(namespace, listOptions, onGetWatchError),
		)

	newWatch.SetFilters([]watch.EventType{watch.Added, watch.Modified})


	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := k.CoreClient.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		t.Error(err)
	}

	newWatch.ListenWatcher(func(event watch.Event) {
		t.Logf("New Event Received %+v", event)
	})
}

