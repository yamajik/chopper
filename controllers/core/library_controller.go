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
	if _, err := r.Resource().Status().Update(ctx, lib, func() error {
		lib.Status.Data = len(lib.Spec.Data) + len(lib.Spec.BinaryData)
		hash, err := lib.Hash()
		if err != nil {
			return err
		}
		lib.Status.Hash = hash
		return nil
	}); err != nil {
		return err
	}

	var rts corev1.RuntimeList
	if _, err := r.Resource().List(ctx, &rts, client.InNamespace(lib.Namespace), client.HasLabels{fmt.Sprintf("kess-lib-%s", lib.Name)}); err != nil {
		return err
	}
	for i := range rts.Items {
		rt := rts.Items[i]
		if _, err := r.Resource().Status().Update(ctx, &rt, func() error {
			rt.Status.Functions = maps.MergeString(rt.Status.Libraries, map[string]string{
				lib.Name: lib.Status.Hash,
			})
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *LibraryReconciler) applyExternalResources(ctx context.Context, lib *corev1.Library) error {
	prevHash := lib.Status.Hash
	hash, err := lib.Hash()
	if err != nil {
		return err
	}
	if hash == prevHash {
		return nil
	}

	var (
		cm           = lib.ConfigMap(hash)
		patchOptions = client.PatchOptions{FieldManager: corev1.FieldManager, Force: pointer.BoolPtr(true)}
	)

	ctrl.SetControllerReference(lib, &cm, r.Scheme)
	if _, err := r.Resource().Patch(ctx, &cm, client.Apply, &patchOptions); err != nil {
		return err
	}

	var cms apiv1.ConfigMapList
	if _, err := r.Resource().List(ctx, &cms, client.InNamespace(lib.Namespace), client.MatchingLabels(lib.Labels())); err != nil {
		return err
	}
	for i := range cms.Items {
		item := cms.Items[i]
		if cm.Name != item.Name {
			if _, err := r.Resource().Delete(ctx, &item); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *LibraryReconciler) deleteExternalResources(ctx context.Context, lib *corev1.Library) error {
	var cm apiv1.ConfigMap
	if _, err := r.Resource().GetAndDelete(ctx, lib.NamespacedName(lib.ConfigMapName(lib.Status.Hash)), &cm); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// SetupWithManager bulabula
func (r *LibraryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Library{}).
		Complete(r)
}
