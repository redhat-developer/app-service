package watcher

import (
	"github.com/redhat-developer/app-service/kubeclient"
	"k8s.io/apimachinery/pkg/watch"
)

type Watch struct {
	Client *kubeclient.KubeClient
	ResultStream chan watch.Event
	Namespace string
	Watchers []watch.Interface
	WatchFilters []watch.EventType
}

func NewWatch(namespace string, kc *kubeclient.KubeClient, watchers ...watch.Interface) *Watch {
	w := new(Watch)
	w.Client = kc
	w.ResultStream = make(chan watch.Event)
	w.Namespace = namespace
	w.Watchers = watchers
	return w
}

func (w *Watch) SetFilters(filters []watch.EventType) *Watch {
	w.WatchFilters = filters
	return w
}

func (w Watch) StartWatcher() {
	for _, v := range w.Watchers {
		go sendToChannel(v, w.ResultStream)
	}
}

func (w Watch) ListenWatcher(onEvent func(obj watch.Event)) {
	for {
		obj := <- w.ResultStream
		for _, v := range w.WatchFilters {
			if v == obj.Type {
				onEvent(obj)
			}
		}
	}
}

func (w Watch) StopWatch() {
	for _, v := range w.Watchers {
		v.Stop()
	}
}

func sendToChannel(w watch.Interface, ch chan watch.Event)  {
	for {
		v := <- w.ResultChan()
		ch <- v
	}
}
