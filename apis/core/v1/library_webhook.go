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
var librarylog = logf.Log.WithName("library-resource")

// SetupWebhookWithManager bulabula
func (r *Library) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-core-kess-io-v1-library,mutating=true,failurePolicy=fail,groups=core.kess.io,resources=libraries,verbs=create;update,versions=v1,name=mlibrary.kb.io,admissionReviewVersions=v1,sideEffects=None

var _ webhook.Defaulter = &Library{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Library) Default() {
	librarylog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	r.ObjectMeta.Labels = utilsmaps.MergeString(r.Labels(), r.ObjectMeta.Labels)
}

// DefaultStatus implements webhook.Defaulter so a webhook will be registered for the type
func (r *Library) DefaultStatus() {
	librarylog.Info("default status", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-core-kess-io-v1-library,mutating=false,failurePolicy=fail,groups=core.kess.io,resources=libraries,versions=v1,name=vlibrary.kb.io,admissionReviewVersions=v1,sideEffects=None

var _ webhook.Validator = &Library{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Library) ValidateCreate() error {
	librarylog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Library) ValidateUpdate(old runtime.Object) error {
	librarylog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Library) ValidateDelete() error {
	librarylog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
