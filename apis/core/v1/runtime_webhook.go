/*
Copyright 2021 yamajik.

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
	utilsmaps "github.com/yamajik/kess/utils/maps"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var runtimelog = logf.Log.WithName("runtime-resource")

// SetupWebhookWithManager bulabula
func (r *Runtime) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-core-kess-io-v1-runtime,mutating=true,failurePolicy=fail,groups=core.kess.io,resources=runtimes,verbs=create;update,versions=v1,name=mruntime.kb.io,admissionReviewVersions=v1,sideEffects=None

var _ webhook.Defaulter = &Runtime{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Runtime) Default() {
	runtimelog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	r.ObjectMeta.Labels = utilsmaps.MergeString(r.Labels(), r.FunctionLabels(), r.LibraryLabels(), r.ObjectMeta.Labels)
}

// DefaultStatus implements webhook.Defaulter so a webhook will be registered for the type
func (r *Runtime) DefaultStatus() {
	runtimelog.Info("default status", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Status.Ready == "" {
		r.Status.Ready = DefaultReady
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-core-kess-io-v1-runtime,mutating=false,failurePolicy=fail,groups=core.kess.io,resources=runtimes,versions=v1,name=vruntime.kb.io,admissionReviewVersions=v1,sideEffects=None

var _ webhook.Validator = &Runtime{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Runtime) ValidateCreate() error {
	runtimelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Runtime) ValidateUpdate(old runtime.Object) error {
	runtimelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Runtime) ValidateDelete() error {
	runtimelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
