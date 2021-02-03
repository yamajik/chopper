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
	"strconv"
	"time"

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
	if _, err := r.Resource().Status().Update(ctx, fn, func() error {
		for name := range fn.Status.RuntimeStatus {
			var rt corev1.Runtime
			if _, err := r.Resource().Get(ctx, fn.RuntimeNamespacedName(name), &rt); err != nil {
				if !apierrors.IsNotFound(err) {
					fn.Status.RuntimeStatus[name] = corev1.DefaultReady
				} else {
					delete(fn.Status.RuntimeStatus, name)
				}
			} else {
				fn.UpdateRuntimeStatus(&rt)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *FunctionReconciler) applyExternalResources(ctx context.Context, fn *corev1.Function) error {
	var cm = fn.ConfigMap()

	result, err := r.Resource().CreateOrUpdate(ctx, &cm, func() error {
		ctrl.SetControllerReference(fn, &cm, r.Scheme)
		return nil
	})
	if err != nil {
		return err
	}

	if result != operations.ResultNone {
		if err := r.rollUpdateRuntime(ctx, fn); err != nil {
			return err
		}
	}

	return nil
}

func (r *FunctionReconciler) deleteExternalResources(ctx context.Context, fn *corev1.Function) error {
	var cm apiv1.ConfigMap

	if _, err := r.Resource().GetAndDelete(ctx, fn.NamespacedName(fn.ConfigMapName()), &cm); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (r *FunctionReconciler) rollUpdateRuntime(ctx context.Context, fn *corev1.Function) error {
	var versionConfig = strconv.Itoa(time.Now().Nanosecond())

	for name := range fn.Status.RuntimeStatus {
		var rt corev1.Runtime
		if _, err := r.Resource().Get(ctx, fn.RuntimeNamespacedName(name), &rt); err != nil {
			continue
		}
		if _, err := r.Resource().Update(ctx, &rt, func() error {
			if rt.Annotations == nil {
				rt.Annotations = make(map[string]string)
			}
			rt.Annotations[corev1.VersionConfig] = versionConfig
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager bulabula
func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Function{}).
		Complete(r)
}
