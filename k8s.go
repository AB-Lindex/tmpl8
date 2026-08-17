package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildK8sClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		).ClientConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

func loadK8s(fn string) ([]entry, error) {
	return loadK8sWithClient(fn, nil)
}

func loadK8sWithClient(fn string, clientset kubernetes.Interface) ([]entry, error) {
	if args.Verbose {
		log.Info().Msgf("loading k8s-config '%s'...", fn)
	}

	parts := strings.Split(fn, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("k8s-format is 'k8s:namespace/configname', not 'k8s:%s'", fn)
	}
	namespace := parts[0]
	name := parts[1]

	if clientset == nil {
		var err error
		clientset, err = buildK8sClient()
		if err != nil {
			return nil, err
		}
	}

	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap %s/%s: %w", namespace, name, err)
	}

	var result []entry
	for key, data := range cm.Data {
		result = append(result, entry{fmt.Sprintf("k8s:%s/%s/%s", namespace, name, key), data})
	}

	return result, nil
}
