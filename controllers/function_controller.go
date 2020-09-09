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

// FunctionReconciler reconciles a Function object
type FunctionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	ops operations.ResourceOperationsInterface
}

// Resource bulabula
func (r *FunctionReconciler) Resource() operations.ResourceOperationsInterface {
	if r.ops == nil {
		r.ops = operations.NewResourceOperations(r.Client)
	}
	return r.ops
}

// +kubebuilder:rbac:groups=core.kess.io,resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kess.io,resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.kess.io,resources=runtimes,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kess.io,resources=runtimes/status,verbs=update;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps/status,verbs=update;patch

// Reconcile bulabula
func (r *FunctionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("function", req.NamespacedName)

	var fn corev1.Function
	if _, err := r.Resource().Get(ctx, req.NamespacedName, &fn); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if _, err := r.Resource().ApplyDefaultAll(ctx, &fn); err != nil {
		log.Error(err, "unable to set default for function")
		return ctrl.Result{}, err
	}

	if fn.DeletionTimestamp.IsZero() {
		if !r.Resource().ContainsFinalizer(&fn, corev1.Finalizer) {
			if _, err := r.Resource().AddFinalizer(ctx, &fn, corev1.Finalizer); err != nil {
				log.Error(err, "unable to add function finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		if r.Resource().ContainsFinalizer(&fn, corev1.Finalizer) {
			if err := r.deleteExternalResources(ctx, &fn); err != nil {
				log.Error(err, "unable to delete function external resources")
				return ctrl.Result{}, err
			}
			if _, err := r.Resource().RemoveFinalizer(ctx, &fn, corev1.Finalizer); err != nil {
				log.Error(err, "unable to remove function finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.applyExternalResources(ctx, &fn); err != nil {
		log.Error(err, "unable to apply function external resources")
		return ctrl.Result{}, err
	}

	if err := r.applyStatus(ctx, &fn); err != nil {
		log.Error(err, "unable to apply function status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *FunctionReconciler) applyStatus(ctx context.Context, fn *corev1.Function) error {
	_, err := r.Resource().Status().Update(ctx, fn, func() error {
		var rt corev1.Runtime
		r.Get(ctx, fn.RuntimeNamespacedName(), &rt)
		fn.UpdateStatusReady(&rt)
		return nil
	})
	return err
}

func (r *FunctionReconciler) applyExternalResources(ctx context.Context, fn *corev1.Function) error {
	var cm = fn.ConfigMap()

	if _, err := r.Resource().CreateOrUpdate(ctx, &cm, func() error {
		fn.SetConfigMap(&cm)
		return nil
	}); err != nil {
		return err
	}

	if err := r.applyRuntimeStatusFunctions(ctx, fn); err != nil {
		return err
	}

	return nil
}

func (r *FunctionReconciler) deleteExternalResources(ctx context.Context, fn *corev1.Function) error {
	var cm apiv1.ConfigMap

	if _, err := r.Resource().Get(ctx, fn.ConfigMapNamespacedName(), &cm); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	fn.UnsetConfigMap(&cm)
	if len(cm.Data) == 0 && len(cm.BinaryData) == 0 {
		if err := r.deleteRuntimeStatusFunctions(ctx, fn); err != nil {
			return err
		}
		if _, err := r.Resource().Delete(ctx, &cm); err != nil {
			return err
		}
	} else {
		if _, err := r.Resource().Update(ctx, &cm, func() error {
			fn.UnsetConfigMap(&cm)
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *FunctionReconciler) applyRuntimeStatusFunctions(ctx context.Context, fn *corev1.Function) error {
	var rt corev1.Runtime

	if _, err := r.Resource().Get(ctx, fn.RuntimeNamespacedName(), &rt); err != nil {
		if apierrors.IsNotFound(err) {
			r.Resource().Status().Update(ctx, fn, func() error {
				fn.UpdateStatusReady(&rt)
				return nil
			})
		}
		return err
	}

	if _, err := r.Resource().Update(ctx, fn, func() error {
		return ctrl.SetControllerReference(&rt, fn, r.Scheme)
	}); err != nil {
		return err
	}

	if _, err := r.Resource().Status().Update(ctx, &rt, func() error {
		rt.UpdateStatusFunctions(fn)
		return nil
	}); err != nil {
		return err
	}

	// TODO: hot upgrade runtime deployment

	return nil
}

func (r *FunctionReconciler) deleteRuntimeStatusFunctions(ctx context.Context, fn *corev1.Function) error {
	var runtime corev1.Runtime

	if _, err := r.Resource().Get(ctx, fn.RuntimeNamespacedName(), &runtime); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	if _, err := r.Resource().Status().Update(ctx, &runtime, func() error {
		runtime.DeleteStatusFunctions(fn)
		return nil
	}); err != nil {
		return err
	}

	// TODO: hot upgrade runtime deployment

	return nil
}

// SetupWithManager bulabula
func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Function{}).
		Complete(r)
}
