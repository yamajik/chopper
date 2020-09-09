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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "github.com/yamajik/kess/api/v1"
	"github.com/yamajik/kess/controllers/operations"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

// RuntimeReconciler reconciles a Runtime object
type RuntimeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	ops operations.ResourceOperationsInterface
}

// Resource bulabula
func (r *RuntimeReconciler) Resource() operations.ResourceOperationsInterface {
	if r.ops == nil {
		r.ops = operations.NewResourceOperations(r.Client)
	}
	return r.ops
}

// +kubebuilder:rbac:groups=core.kess.io,resources=runtimes,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kess.io,resources=runtimes/status,verbs=update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=update;patch

// Reconcile runtime
func (r *RuntimeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("runtime", req.NamespacedName)

	var rt corev1.Runtime
	if _, err := r.Resource().Get(ctx, req.NamespacedName, &rt); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if _, err := r.Resource().ApplyDefaultAll(ctx, &rt); err != nil {
		log.Error(err, "unable to set default for runtime")
		return ctrl.Result{}, err
	}

	if rt.DeletionTimestamp.IsZero() {
		if !r.Resource().ContainsFinalizer(&rt, corev1.Finalizer) {
			if _, err := r.Resource().AddFinalizer(ctx, &rt, corev1.Finalizer); err != nil {
				log.Error(err, "unable to add runtime finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		if r.Resource().ContainsFinalizer(&rt, corev1.Finalizer) {
			if err := r.deleteExternalResources(ctx, &rt); err != nil {
				log.Error(err, "unable to delete runtime external resources")
				return ctrl.Result{}, err
			}
			if _, err := r.Resource().RemoveFinalizer(ctx, &rt, corev1.Finalizer); err != nil {
				log.Error(err, "unable to remove runtime finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.applyExternalResources(ctx, &rt); err != nil {
		log.Error(err, "unable to apply runtime external resources")
		return ctrl.Result{}, err
	}

	if err := r.applyStatus(ctx, &rt); err != nil {
		log.Error(err, "unable to apply runtime status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RuntimeReconciler) applyStatus(ctx context.Context, rt *corev1.Runtime) error {
	var (
		fns         corev1.FunctionList
		libs        corev1.LibraryList
		matchLabels = client.MatchingLabels{"kess-runtime": rt.Name}
	)

	if _, err := r.Resource().Status().Update(ctx, rt, func() error {
		var deploy appsv1.Deployment
		r.Get(ctx, rt.NamespacedName(), &deploy)
		rt.UpdateStatusReady(&deploy)
		return nil
	}); err != nil {
		return err
	}

	if _, err := r.Resource().List(ctx, &fns, matchLabels); err != nil {
		return err
	}
	if _, err := r.Resource().List(ctx, &libs, matchLabels); err != nil {
		return err
	}

	var errors []error
	for _, fn := range fns.Items {
		if _, err := r.Resource().Status().Update(ctx, &fn, func() error {
			fn.UpdateStatusReady(rt)
			return nil
		}); err != nil {
			errors = append(errors, err)
		}
	}
	for _, lib := range libs.Items {
		if _, err := r.Resource().Status().Update(ctx, &lib, func() error {
			lib.UpdateStatusReady(rt)
			return nil
		}); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func (r *RuntimeReconciler) applyExternalResources(ctx context.Context, rt *corev1.Runtime) error {
	var (
		deploy       = rt.Deployment()
		svc          = rt.Service()
		patchOptions = client.PatchOptions{FieldManager: corev1.FieldManager}
	)

	ctrl.SetControllerReference(rt, &deploy, r.Scheme)
	if _, err := r.Resource().Patch(ctx, &deploy, client.Apply, &patchOptions); err != nil {
		return err
	}

	ctrl.SetControllerReference(rt, &svc, r.Scheme)
	if _, err := r.Resource().Patch(ctx, &svc, client.Apply, &patchOptions); err != nil {
		return err
	}

	return nil
}

func (r *RuntimeReconciler) deleteExternalResources(ctx context.Context, rt *corev1.Runtime) error {
	var (
		deploy         appsv1.Deployment
		svc            apiv1.Service
		namespacedName = rt.NamespacedName()
		deleteOptions  = client.DeleteOptions{}
	)

	if _, err := r.Resource().GetAndDelete(ctx, namespacedName, &deploy, &deleteOptions); err != nil {
		return err
	}

	if _, err := r.Resource().GetAndDelete(ctx, namespacedName, &svc, &deleteOptions); err != nil {
		return err
	}

	return nil
}

// SetupWithManager runtime
func (r *RuntimeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Runtime{}).
		Complete(r)
}
