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

// FunctionFile bulabula
type FunctionFile struct {
	// The filename format of function
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="{Version}"
	Name string `json:"name,omitempty"`
}

// FunctionConfigMap bulabula
type FunctionConfigMap struct {
	// The filename format of function
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="fn-{Name}"
	Name string `json:"name,omitempty"`

	// The filename format of function
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/kess/fn/{Name}"
	Mount string `json:"mount,omitempty"`
}

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Optional version of function
	// +kubebuilder:validation:Optional
	Function string `json:"function,omitempty"`

	// Optional version of function
	// +kubebuilder:validation:Optional
	Version string `json:"version,omitempty"`

	// The runtime name of function
	// +kubebuilder:validation:Required
	Runtime string `json:"runtime,omitempty"`

	// The filename format of function
	// +kubebuilder:validation:Optional
	File FunctionFile `json:"file,omitempty"`

	// The filename format of function
	// +kubebuilder:validation:Optional
	ConfigMap FunctionConfigMap `json:"configMap,omitempty"`

	// The string of function
	// +kubebuilder:validation:Optional
	Data string `json:"data,omitempty"`

	// The binary of function
	// +kubebuilder:validation:Optional
	BinaryData []byte `json:"binaryData,omitempty"`
}

// FunctionStatus defines the observed state of Function
type FunctionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Optional ready string of runtime for show
	// +kubebuilder:validation:Optional
	Ready string `json:"ready,omitempty"`
}

// +kubebuilder:resource:categories="kess",shortName="fn"
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.ready`,priority=0
// +kubebuilder:printcolumn:name="Function",type=string,JSONPath=`.spec.function`,priority=0
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,priority=0
// +kubebuilder:printcolumn:name="Runtime",type=string,JSONPath=`.spec.runtime`,priority=0
// +kubebuilder:object:root=true

// Function is the Schema for the functions API
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec,omitempty"`
	Status FunctionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FunctionList contains a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Function{}, &FunctionList{})
}
