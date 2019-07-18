package util

import (
	"fmt"

	"github.com/appscode/go/types"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/portforward"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
)

func WaitUntilFailedNodeAvailable(kubeClient kubernetes.Interface, statefulSet *apps.StatefulSet) error {
	return core_util.WaitUntilPodRunningBySelector(
		kubeClient,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
}

func ForwardPort(
	kubeClient kubernetes.Interface,
	restConfig *rest.Config,
	meta metav1.ObjectMeta, clientPodName string, port int) (*portforward.Tunnel, error) {
	tunnel := portforward.NewTunnel(
		kubeClient.CoreV1().RESTClient(),
		restConfig,
		meta.Namespace,
		clientPodName,
		port,
	)

	if err := tunnel.ForwardPort(); err != nil {
		return nil, err
	}
	return tunnel, nil
}

func GetPods(
	kubeClient kubernetes.Interface,
	meta metav1.ObjectMeta, selector labels.Set) (*core.PodList, error) {
	return kubeClient.CoreV1().Pods(meta.Namespace).List(metav1.ListOptions{
		LabelSelector: selector.String(),
	})
}

func FowardedPodsIPWithTunnel(
	kubeClient kubernetes.Interface, restConfig *rest.Config,
	redis *api.Redis) ([][]string, [][]*portforward.Tunnel, error) {

	var (
		rdAddresses [][]string
		tunnels     [][]*portforward.Tunnel
		err         error
		podName     string
	)
	rdAddresses = make([][]string, int(*redis.Spec.Cluster.Master))
	tunnels = make([][]*portforward.Tunnel, int(*redis.Spec.Cluster.Master))
	for i := 0; i < int(*redis.Spec.Cluster.Master); i++ {
		rdAddresses[i] = make([]string, int(*redis.Spec.Cluster.Replicas)+1)
		tunnels[i] = make([]*portforward.Tunnel, int(*redis.Spec.Cluster.Replicas)+1)
		for j := 0; j <= int(*redis.Spec.Cluster.Replicas); j++ {
			podName = fmt.Sprintf("%s-shard%d-%d", redis.Name, i, j)
			if tunnels[i][j], err = ForwardPort(kubeClient, restConfig, redis.ObjectMeta, podName, 6379); err != nil {
				return nil, nil, err
			}
			rdAddresses[i][j] = fmt.Sprintf("%d", tunnels[i][j].Local)
		}
	}

	return rdAddresses, tunnels, nil
}
