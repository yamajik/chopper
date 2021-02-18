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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "github.com/yamajik/kess/apis/core/v1"
	"github.com/yamajik/kess/controllers/operations"
	"github.com/yamajik/kess/utils/maps"
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
		fn.Status.Data = len(fn.Spec.Data) + len(fn.Spec.BinaryData)
		hash, err := fn.Hash()
		if err != nil {
			return err
		}
		fn.Status.Hash = hash
		return nil
	}); err != nil {
		return err
	}

	var rts corev1.RuntimeList
	if _, err := r.Resource().List(ctx, &rts, client.InNamespace(fn.Namespace), client.HasLabels{fmt.Sprintf("kess-fn-%s", fn.Name)}); err != nil {
		return err
	}
	for i := range rts.Items {
		rt := rts.Items[i]
		if _, err := r.Resource().Status().Update(ctx, &rt, func() error {
			rt.Status.Functions = maps.MergeString(rt.Status.Functions, map[string]string{
				fn.Name: fn.Status.Hash,
			})
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *FunctionReconciler) applyExternalResources(ctx context.Context, fn *corev1.Function) error {
	prevHash := fn.Status.Hash
	hash, err := fn.Hash()
	if err != nil {
		return err
	}
	if hash == prevHash {
		return nil
	}

	var (
		cm           = fn.ConfigMap(hash)
		patchOptions = client.PatchOptions{FieldManager: corev1.FieldManager, Force: pointer.BoolPtr(true)}
	)

	ctrl.SetControllerReference(fn, &cm, r.Scheme)
	if _, err := r.Resource().Patch(ctx, &cm, client.Apply, &patchOptions); err != nil {
		return err
	}

	var cms apiv1.ConfigMapList
	if _, err := r.Resource().List(ctx, &cms, client.InNamespace(fn.Namespace), client.MatchingLabels(fn.Labels())); err != nil {
		return err
	}
	for i := range cms.Items {
		item := cms.Items[i]
		if item.Name != cm.Name {
			if _, err := r.Resource().Delete(ctx, &item); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *FunctionReconciler) deleteExternalResources(ctx context.Context, fn *corev1.Function) error {
	var cm apiv1.ConfigMap
	if _, err := r.Resource().GetAndDelete(ctx, fn.NamespacedName(fn.ConfigMapName(fn.Status.Hash)), &cm); err != nil {
		if !apierrors.IsNotFound(err) {
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
