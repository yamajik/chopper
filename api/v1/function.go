package v1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/xorcare/pointer"
)

// Default bulabula
func (r *Function) Default() {
	var (
		labels       = r.Labels()
		namedVersion = r.NamedVersion()
	)

	if r.ObjectMeta.Labels == nil {
		r.ObjectMeta.Labels = make(map[string]string)
	}
	for k, v := range labels {
		r.ObjectMeta.Labels[k] = v
	}

	if r.Spec.Function == "" {
		r.Spec.Function = namedVersion.Name
	}
	if r.Spec.Version == "" {
		r.Spec.Version = namedVersion.Version
	}
}

// DefaultStatus bulabula
func (r *Function) DefaultStatus() {
	if r.Status.Ready == "" {
		r.Status.Ready = DefaultReady
	}
}

// NamedVersion bulabula
func (r *Function) NamedVersion() NamedVersion {
	return NamedVersionFromString(r.Name)
}

// RuntimeConfigMap bulabula
func (r *Function) RuntimeConfigMap() RuntimeConfigMap {
	namedVersion := r.NamedVersion()
	return RuntimeConfigMap{
		Name:  namedVersion.Format(r.Spec.ConfigMap.Name),
		Mount: namedVersion.Format(r.Spec.ConfigMap.Mount),
	}
}

// Labels bulabula
func (r *Function) Labels() map[string]string {
	return map[string]string{
		"kess-type":     TypeFunction,
		"kess-function": r.Spec.Function,
		"kess-version":  r.Spec.Version,
		"kess-runtime":  r.Spec.Runtime,
	}
}

// RuntimeNamespacedName bulabula
func (r *Function) RuntimeNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.Spec.Runtime,
		Namespace: r.Namespace,
	}
}

// ConfigMap bulabula
func (r *Function) ConfigMap() apiv1.ConfigMap {
	labels := r.Labels()

	configmap := apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.RuntimeConfigMap().Name,
			Namespace: r.Namespace,
			Labels:    labels,
		},
		Immutable: pointer.Bool(true),
	}

	r.SetConfigMap(&configmap)

	return configmap
}

// ConfigMapNamespacedName bulabula
func (r *Function) ConfigMapNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.RuntimeConfigMap().Name,
		Namespace: r.Namespace,
	}
}

// SetConfigMap bulabula
func (r *Function) SetConfigMap(out *apiv1.ConfigMap) {
	key := r.NamedVersion().Format(r.Spec.File.Name)
	if r.Spec.Data != "" {
		if out.Data == nil {
			out.Data = make(map[string]string)
		}
		out.Data[key] = r.Spec.Data
	}
	if len(r.Spec.BinaryData) > 0 {
		if out.BinaryData == nil {
			out.BinaryData = make(map[string][]byte)
		}
		out.BinaryData[key] = r.Spec.BinaryData
	}
}

// UnsetConfigMap bulabula
func (r *Function) UnsetConfigMap(out *apiv1.ConfigMap) {
	key := r.NamedVersion().Format(r.Spec.File.Name)
	if out.Data != nil {
		delete(out.Data, key)
	}
	if out.BinaryData != nil {
		delete(out.BinaryData, key)
	}
}

// UpdateStatusReady bulabula
func (r *Function) UpdateStatusReady(rt *Runtime) {
	r.Status.Ready = rt.Status.Ready
}
