package v1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Default bulabula
func (r *Library) Default() {
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

	if r.Spec.Library == "" {
		r.Spec.Library = namedVersion.Name
	}
	if r.Spec.Version == "" {
		r.Spec.Version = namedVersion.Version
	}
}

// DefaultStatus bulabula
func (r *Library) DefaultStatus() {
	if r.Status.Ready == "" {
		r.Status.Ready = DefaultReady
	}
}

// NamedVersion bulabula
func (r *Library) NamedVersion() NamedVersion {
	return NamedVersionFromString(r.Name)
}

// Labels bulabula
func (r *Library) Labels() map[string]string {
	return map[string]string{
		"kess-type":    TypeLibrary,
		"kess-library": r.Spec.Library,
		"kess-version": r.Spec.Version,
		"kess-runtime": r.Spec.Runtime,
	}
}

// RuntimeNamespacedName bulabula
func (r *Library) RuntimeNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.Spec.Runtime,
		Namespace: r.Namespace,
	}
}

// RuntimeConfigMap bulabula
func (r *Library) RuntimeConfigMap() RuntimeConfigMap {
	namedVersion := r.NamedVersion()
	return RuntimeConfigMap{
		Name:  namedVersion.Format(r.Spec.ConfigMap.Name),
		Mount: namedVersion.Format(r.Spec.ConfigMap.Mount),
	}
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
			Name:      r.RuntimeConfigMap().Name,
			Namespace: r.Namespace,
			Labels:    labels,
		},
		Data:       r.Spec.Data,
		BinaryData: r.Spec.BinaryData,
		// Immutable:  pointer.Bool(true),
	}

	return configmap
}

// ConfigMapNamespacedName bulabula
func (r *Library) ConfigMapNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.RuntimeConfigMap().Name,
		Namespace: r.Namespace,
	}
}

// SetConfigMap bulabula
func (r *Library) SetConfigMap(out *apiv1.ConfigMap) {
	out.Data = r.Spec.Data
	out.BinaryData = r.Spec.BinaryData
}

// UnsetConfigMap bulabula
func (r *Library) UnsetConfigMap(out *apiv1.ConfigMap) {
	out.Data = nil
	out.BinaryData = nil
}

// UpdateStatusReady bulabula
func (r *Library) UpdateStatusReady(rt *Runtime) {
	r.Status.Ready = rt.Status.Ready
}

// AddFinalizer bulabula
func (r *Library) AddFinalizer(finalizer string) error {
	controllerutil.AddFinalizer(r, finalizer)
	return nil
}

// RemoveFinalizer bulabula
func (r *Library) RemoveFinalizer(finalizer string) error {
	controllerutil.RemoveFinalizer(r, finalizer)
	return nil
}
