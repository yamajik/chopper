package v1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/yamajik/kess/utils/strings"
)

// Default bulabula
func (r *Library) Default() {
	var (
		labels = r.Labels()
	)

	if r.ObjectMeta.Labels == nil {
		r.ObjectMeta.Labels = make(map[string]string)
	}
	for k, v := range labels {
		r.ObjectMeta.Labels[k] = v
	}
}

// DefaultStatus bulabula
func (r *Library) DefaultStatus() {
	r.Status.ConfigMap = r.ConfigMapName()

	if r.Status.RuntimeStatus == nil {
		r.Status.RuntimeStatus = make(map[string]string)
	}
}

// Labels bulabula
func (r *Library) Labels() map[string]string {
	return map[string]string{
		"kess-type":    TypeLibrary,
		"kess-library": r.Name,
	}
}

// NamespacedName bulabula
func (r *Library) NamespacedName(name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: r.Namespace,
	}
}

// ConfigMapName bulabula
func (r *Library) ConfigMapName() string {
	m := map[string]interface{}{
		"Name": r.Name,
	}
	return strings.Format(r.Spec.ConfigMap, m)
}

// ConfigMap bulabula
func (r *Library) ConfigMap() apiv1.ConfigMap {
	labels := r.Labels()

	configmap := apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ConfigMapName(),
			Namespace: r.Namespace,
			Labels:    labels,
		},
		Data:       r.Spec.Data,
		BinaryData: r.Spec.BinaryData,
	}

	return configmap
}

// RuntimeNamespacedName bulabula
func (r *Library) RuntimeNamespacedName(name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: r.Namespace,
	}
}

// UpdateRuntimeStatus bulabula
func (r *Library) UpdateRuntimeStatus(rt *Runtime) {
	if r.Status.RuntimeStatus == nil {
		r.Status.RuntimeStatus = make(map[string]string)
	}
	r.Status.RuntimeStatus[rt.Name] = rt.Status.Ready
}
