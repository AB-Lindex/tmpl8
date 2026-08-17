//go:build integration

package main

import (
	"context"
	"sort"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testNamespace = "default"
	testCMName    = "tmpl8-test"
)

// testConfigMapData is the expected content of the deployed ConfigMap.
// It must match tests/k8s/tmpl8-test-configmap.yaml.
var testConfigMapData = map[string]string{
	"greeting":  "hello from configmap",
	"count":     "42",
	"multiline": "line one\nline two\nline three\n",
}

func TestLoadK8s_Integration(t *testing.T) {
	client, err := buildK8sClient()
	if err != nil {
		t.Fatalf("failed to build k8s client (is a cluster accessible?): %v", err)
	}

	ctx := context.Background()

	// Ensure the ConfigMap exists; create or update it.
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testCMName,
			Namespace: testNamespace,
		},
		Data: testConfigMapData,
	}

	existing, err := client.CoreV1().ConfigMaps(testNamespace).Get(ctx, testCMName, metav1.GetOptions{})
	if err != nil {
		// Not found – create it.
		_, err = client.CoreV1().ConfigMaps(testNamespace).Create(ctx, cm, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create test ConfigMap %s/%s: %v", testNamespace, testCMName, err)
		}
		t.Logf("created ConfigMap %s/%s", testNamespace, testCMName)
		t.Cleanup(func() {
			_ = client.CoreV1().ConfigMaps(testNamespace).Delete(ctx, testCMName, metav1.DeleteOptions{})
			t.Logf("deleted ConfigMap %s/%s", testNamespace, testCMName)
		})
	} else {
		// Already exists – update in case data differs.
		existing.Data = testConfigMapData
		_, err = client.CoreV1().ConfigMaps(testNamespace).Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			t.Fatalf("failed to update test ConfigMap %s/%s: %v", testNamespace, testCMName, err)
		}
		t.Logf("updated existing ConfigMap %s/%s", testNamespace, testCMName)
	}

	// Call loadK8sWithClient with the real client.
	entries, err := loadK8sWithClient(testNamespace+"/"+testCMName, client)
	if err != nil {
		t.Fatalf("loadK8sWithClient returned error: %v", err)
	}

	if len(entries) != len(testConfigMapData) {
		t.Fatalf("expected %d entries, got %d", len(testConfigMapData), len(entries))
	}

	// Build a map for easy lookup (order from range over map is non-deterministic).
	got := make(map[string]string, len(entries))
	for _, e := range entries {
		// entry name is "k8s:namespace/configname/key"
		got[e.name] = e.data
	}

	keys := make([]string, 0, len(testConfigMapData))
	for k := range testConfigMapData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		wantName := "k8s:" + testNamespace + "/" + testCMName + "/" + key
		wantData := testConfigMapData[key]
		gotData, ok := got[wantName]
		if !ok {
			t.Errorf("missing entry %q in result", wantName)
			continue
		}
		if gotData != wantData {
			t.Errorf("entry %q: expected data %q, got %q", wantName, wantData, gotData)
		}
	}
}

func TestLoadK8s_InvalidFormat(t *testing.T) {
	_, err := loadK8sWithClient("no-slash-here", nil)
	if err == nil {
		t.Fatal("expected an error for invalid format, got nil")
	}
}
