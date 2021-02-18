package v1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/yamajik/kess/utils/hash"
	"github.com/yamajik/kess/utils/strings"
)

// Labels bulabula
func (r *Function) Labels() map[string]string {
	return map[string]string{
		"kess-type": TypeFunction,
		"kess-fn":   r.Name,
	}
}

// HashLabels bulabula
func (r *Function) HashLabels(hash string) map[string]string {
	return map[string]string{
		"kess-hash": hash,
	}
}

// RuntimeStatus bulabula
func (r *Function) RuntimeStatus() map[string]string {
	return map[string]string{}
}

// NamespacedName bulabula
func (r *Function) NamespacedName(name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: r.Namespace,
	}
}

// ConfigMapName bulabula
func (r *Function) ConfigMapName(hash string) string {
	m := map[string]interface{}{
		"Name": r.Name,
		"Hash": hash,
	}
	return strings.Format(r.Spec.ConfigMap, m)
}

// ConfigMap bulabula
func (r *Function) ConfigMap(hash string) apiv1.ConfigMap {
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
func (r *Function) RuntimeNamespacedName(name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: r.Namespace,
	}
}

// ConfigMapNamespacedName bulabula
func (r *Function) ConfigMapNamespacedName(hash string) types.NamespacedName {
	return types.NamespacedName{
		Name:      r.ConfigMapName(hash),
		Namespace: r.Namespace,
	}
}

// Hash bulabula
func (r *Function) Hash() (string, error) {
	var m = map[string]interface{}{
		"Data":       r.Spec.Data,
		"BinaryData": r.Spec.BinaryData,
	}

	return hash.FromMap(m)
}
