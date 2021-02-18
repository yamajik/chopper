package v1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/yamajik/kess/utils/hash"
	"github.com/yamajik/kess/utils/strings"
)

// Labels bulabula
func (r *Library) Labels() map[string]string {
	return map[string]string{
		"kess-type": TypeLibrary,
		"kess-lib":  r.Name,
	}
}

// HashLabels bulabula
func (r *Library) HashLabels(hash string) map[string]string {
	return map[string]string{
		"kess-hash": hash,
	}
}

// RuntimeStatus bulabula
func (r *Library) RuntimeStatus() map[string]string {
	return map[string]string{}
}

// NamespacedName bulabula
func (r *Library) NamespacedName(name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: r.Namespace,
	}
}

// ConfigMapName bulabula
func (r *Library) ConfigMapName(hash string) string {
	m := map[string]interface{}{
		"Name": r.Name,
		"Hash": hash,
	}
	return strings.Format(r.Spec.ConfigMap, m)
}

// ConfigMap bulabula
func (r *Library) ConfigMap(hash string) apiv1.ConfigMap {
	labels := r.Labels()

	configmap := apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ConfigMapName(hash),
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

// Hash bulabula
func (r *Library) Hash() (string, error) {
	var m = map[string]interface{}{
		"Data":       r.Spec.Data,
		"BinaryData": r.Spec.BinaryData,
	}

	return hash.FromMap(m)
}
