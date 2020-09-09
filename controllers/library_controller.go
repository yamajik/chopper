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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "github.com/yamajik/kess/api/v1"
	"github.com/yamajik/kess/controllers/operations"
)

// LibraryReconciler reconciles a Library object
type LibraryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	ops operations.ResourceOperationsInterface
}

// Resource bulabula
func (r *LibraryReconciler) Resource() operations.ResourceOperationsInterface {
	if r.ops == nil {
		r.ops = operations.NewResourceOperations(r.Client)
	}
	return r.ops
}

// +kubebuilder:rbac:groups=core.kess.io,resources=libraries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kess.io,resources=libraries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.kess.io,resources=runtimes,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kess.io,resources=runtimes/status,verbs=update;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps/status,verbs=update;patch

// Reconcile bulabula
func (r *LibraryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("library", req.NamespacedName)

	var lib corev1.Library
	if _, err := r.Resource().Get(ctx, req.NamespacedName, &lib); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if _, err := r.Resource().ApplyDefaultAll(ctx, &lib); err != nil {
		log.Error(err, "unable to set default for library")
		return ctrl.Result{}, err
	}

	if lib.DeletionTimestamp.IsZero() {
		if !r.Resource().ContainsFinalizer(&lib, corev1.Finalizer) {
			if _, err := r.Resource().AddFinalizer(ctx, &lib, corev1.Finalizer); err != nil {
				log.Error(err, "unable to add library finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		if r.Resource().ContainsFinalizer(&lib, corev1.Finalizer) {
			if err := r.deleteExternalResources(ctx, &lib); err != nil {
				log.Error(err, "unable to delete library external resources")
				return ctrl.Result{}, err
			}
			if _, err := r.Resource().RemoveFinalizer(ctx, &lib, corev1.Finalizer); err != nil {
				log.Error(err, "unable to remove library finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.applyExternalResources(ctx, &lib); err != nil {
		log.Error(err, "unable to apply library external resources")
		return ctrl.Result{}, err
	}

	if err := r.applyStatus(ctx, &lib); err != nil {
		log.Error(err, "unable to apply library status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *LibraryReconciler) applyStatus(ctx context.Context, lib *corev1.Library) error {
	_, err := r.Resource().Status().Update(ctx, lib, func() error {
		var rt corev1.Runtime
		r.Get(ctx, lib.RuntimeNamespacedName(), &rt)
		lib.UpdateStatusReady(&rt)
		return nil
	})
	return err
}

func (r *LibraryReconciler) applyExternalResources(ctx context.Context, lib *corev1.Library) error {
	var (
		cm           = lib.ConfigMap()
		patchOptions = client.PatchOptions{FieldManager: corev1.FieldManager}
	)

	ctrl.SetControllerReference(lib, &cm, r.Scheme)
	if _, err := r.Resource().Patch(ctx, &cm, client.Apply, &patchOptions); err != nil {
		return err
	}

	if err := r.applyRuntimeStatusLibraries(ctx, lib); err != nil {
		return err
	}

	return nil
}

func (r *LibraryReconciler) deleteExternalResources(ctx context.Context, lib *corev1.Library) error {
	var (
		cm            apiv1.ConfigMap
		deleteOptions = client.DeleteOptions{}
	)

	if err := r.deleteRuntimeStatusLibraries(ctx, lib); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	if _, err := r.Resource().GetAndDelete(ctx, lib.ConfigMapNamespacedName(), &cm, &deleteOptions); err != nil {
		return err
	}

	return nil
}

func (r *LibraryReconciler) applyRuntimeStatusLibraries(ctx context.Context, lib *corev1.Library) error {
	var rt corev1.Runtime
	if _, err := r.Resource().Get(ctx, lib.RuntimeNamespacedName(), &rt); err != nil {
		if apierrors.IsNotFound(err) {
			r.Resource().Status().Update(ctx, lib, func() error {
				lib.UpdateStatusReady(&rt)
				return nil
			})
		}
		return err
	}

	if _, err := r.Resource().Update(ctx, lib, func() error {
		return ctrl.SetControllerReference(&rt, lib, r.Scheme)
	}); err != nil {
		return err
	}

	if _, err := r.Resource().Status().Update(ctx, &rt, func() error {
		rt.UpdateStatusLibraries(lib)
		return nil
	}); err != nil {
		return err
	}

	// TODO: hot update runtime deployment

	return nil
}

func (r *LibraryReconciler) deleteRuntimeStatusLibraries(ctx context.Context, lib *corev1.Library) error {
	var rt corev1.Runtime

	if _, err := r.Resource().Get(ctx, lib.RuntimeNamespacedName(), &rt); err != nil {
		return err
	}

	if _, err := r.Resource().Status().Update(ctx, &rt, func() error {
		rt.DeleteStatusLibraries(lib)
		return nil
	}); err != nil {
		return err
	}

	// TODO: hot upgrade runtime deployment

	return nil
}

// SetupWithManager bulabula
func (r *LibraryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Library{}).
		Complete(r)
}
