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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LibraryConfigMap bulabula
type LibraryConfigMap struct {
	// The filename format of function
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="lib-{Name}-{Version}"
	Name string `json:"name,omitempty"`

	// The filename format of function
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/kess/lib/{Name}-{Version}"
	Mount string `json:"mount,omitempty"`
}

// LibrarySpec defines the desired state of Library
type LibrarySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Optional version of function
	// +kubebuilder:validation:Optional
	Library string `json:"library,omitempty"`

	// Optional version of function
	// +kubebuilder:validation:Optional
	Version string `json:"version,omitempty"`

	// The runtime name of lib
	// +kubebuilder:validation:Required
	Runtime string `json:"runtime,omitempty"`

	// The filename format of lib
	// +kubebuilder:validation:Optional
	ConfigMap LibraryConfigMap `json:"configMap,omitempty"`

	// The string of lib
	// +kubebuilder:validation:Optional
	Data map[string]string `json:"data,omitempty"`

	// The binary of lib
	// +kubebuilder:validation:Optional
	BinaryData map[string][]byte `json:"binaryData,omitempty"`
}

// LibraryStatus defines the observed state of Library
type LibraryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Optional ready string of runtime for show
	// +kubebuilder:validation:Optional
	Ready string `json:"ready,omitempty"`
}

// +kubebuilder:resource:categories="kess",shortName="lib",singular="library"
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.ready`,priority=0
// +kubebuilder:printcolumn:name="Library",type=string,JSONPath=`.spec.library`,priority=0
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,priority=0
// +kubebuilder:printcolumn:name="Runtime",type=string,JSONPath=`.spec.runtime`,priority=0
// +kubebuilder:object:root=true

// Library is the Schema for the Libraries API
type Library struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LibrarySpec   `json:"spec,omitempty"`
	Status LibraryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LibraryList contains a list of Library
type LibraryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Library `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Library{}, &LibraryList{})
}
