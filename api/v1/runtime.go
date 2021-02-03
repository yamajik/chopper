/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"strconv"

	utilsstrings "github.com/yamajik/kess/utils/strings"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Default bulabula
func (r *Runtime) Default() {
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
func (r *Runtime) DefaultStatus() {
	if r.Status.Ready == "" {
		r.Status.Ready = DefaultReady
	}
}

// Labels bulabula
func (r *Runtime) Labels() map[string]string {
	return map[string]string{
		"kess-type":    TypeRuntime,
		"kess-runtime": r.Name,
	}
}

// NamespacedName bulabula
func (r *Runtime) NamespacedName(name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: r.Namespace,
	}
}

// Deployment bulabula
func (r *Runtime) Deployment(volumes []apiv1.Volume, mounts []apiv1.VolumeMount) appsv1.Deployment {
	labels := r.Labels()

	port := apiv1.ContainerPort{
		Name:          r.Spec.PortName,
		ContainerPort: r.Spec.Port,
		Protocol:      apiv1.ProtocolTCP,
	}

	container := apiv1.Container{
		Name:         r.Name,
		Image:        r.Spec.Image,
		Command:      r.Spec.Command,
		Ports:        []apiv1.ContainerPort{port},
		VolumeMounts: mounts,
	}

	template := apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.Name,
			Namespace:   r.Namespace,
			Labels:      labels,
			Annotations: r.Annotations,
		},
		Spec: apiv1.PodSpec{
			Volumes:    volumes,
			Containers: []apiv1.Container{container},
		},
	}

	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: template,
			Replicas: r.Spec.Replicas,
		},
	}

	return deployment
}

// Service bulabula
func (r *Runtime) Service() apiv1.Service {
	labels := r.Labels()

	port := apiv1.ServicePort{
		Name:       r.Spec.PortName,
		Port:       r.Spec.Port,
		TargetPort: intstr.FromInt(int(r.Spec.Port)),
		Protocol:   apiv1.ProtocolTCP,
	}

	service := apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    labels,
		},
		Spec: apiv1.ServiceSpec{
			Selector:  labels,
			ClusterIP: r.Spec.ClusterIP,
			Ports:     []apiv1.ServicePort{port},
		},
	}

	return service
}

// UpdateStatusReady bulabula
func (r *Runtime) UpdateStatusReady(deploy *appsv1.Deployment) {
	r.Status.Ready = utilsstrings.Format(r.Spec.ReadyFormat, map[string]interface{}{
		"Replicas":            strconv.Itoa(int(deploy.Status.Replicas)),
		"UpdatedReplicas":     strconv.Itoa(int(deploy.Status.UpdatedReplicas)),
		"ReadyReplicas":       strconv.Itoa(int(deploy.Status.ReadyReplicas)),
		"AvailableReplicas":   strconv.Itoa(int(deploy.Status.AvailableReplicas)),
		"UnavailableReplicas": strconv.Itoa(int(deploy.Status.UnavailableReplicas)),
	})
}
