package testingutil

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	odolabels "github.com/redhat-developer/odo/pkg/labels"
)

// CreateFakePod creates a fake pod with the given pod name and component name
func CreateFakePod(componentName, podName string) *corev1.Pod {
	fakePod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: odolabels.GetLabels(componentName, "app", "", odolabels.ComponentDevMode, false),
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	return fakePod
}
