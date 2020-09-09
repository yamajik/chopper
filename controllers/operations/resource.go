package operations

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ResourceOperationsInterface bulabula
type ResourceOperationsInterface interface {
	StatusOperationsGetter

	Get(ctx context.Context, key types.NamespacedName, obj runtime.Object) (Result, error)
	List(ctx context.Context, list runtime.Object, opts ...client.ListOption) (Result, error)
	Create(ctx context.Context, obj runtime.Object) (Result, error)
	Update(ctx context.Context, obj runtime.Object, f MutateFn, options ...client.UpdateOption) (Result, error)
	Patch(ctx context.Context, obj runtime.Object, patch client.Patch, options ...client.PatchOption) (Result, error)
	Delete(ctx context.Context, obj runtime.Object, options ...client.DeleteOption) (Result, error)
	CreateOrUpdate(ctx context.Context, obj runtime.Object, f MutateFn, options ...client.UpdateOption) (Result, error)
	GetAndDelete(ctx context.Context, key types.NamespacedName, obj runtime.Object, options ...client.DeleteOption) (Result, error)

	ContainsFinalizer(obj Object, finalizer string) bool
	AddFinalizer(ctx context.Context, obj Object, finalizer string) (Result, error)
	RemoveFinalizer(ctx context.Context, obj Object, finalizer string) (Result, error)
	ApplyDefault(ctx context.Context, obj Object) (Result, error)
	ApplyDefaultAll(ctx context.Context, obj Object) (Result, error)
}

// ResourceOperationsGetter bulabula
type ResourceOperationsGetter interface {
	Resource() ResourceOperationsInterface
}

// ResourceOperations bulabula
type ResourceOperations struct {
	client.Client

	ops StatusOperationsInterface
}

// NewResourceOperations bulabula
func NewResourceOperations(c client.Client) ResourceOperationsInterface {
	return &ResourceOperations{Client: c}
}

// Status bulabula
func (r *ResourceOperations) Status() StatusOperationsInterface {
	if r.ops == nil {
		r.ops = NewStatusOperations(r.Client)
	}
	return r.ops
}

// Get bulabula
func (r *ResourceOperations) Get(ctx context.Context, key types.NamespacedName, obj runtime.Object) (Result, error) {
	if err := r.Client.Get(ctx, key, obj); err != nil {
		return ResultNone, err
	}
	return ResultGot, nil
}

// List bulabula
func (r *ResourceOperations) List(ctx context.Context, list runtime.Object, options ...client.ListOption) (Result, error) {
	if err := r.Client.List(ctx, list, options...); err != nil {
		return ResultNone, err
	}
	return ResultListed, nil
}

// Create bulabula
func (r *ResourceOperations) Create(ctx context.Context, obj runtime.Object) (Result, error) {
	if err := r.Client.Create(ctx, obj); err != nil {
		return ResultNone, err
	}
	return ResultCreated, nil
}

// Update bulabula
func (r *ResourceOperations) Update(ctx context.Context, obj runtime.Object, f MutateFn, options ...client.UpdateOption) (Result, error) {
	key, err := client.ObjectKeyFromObject(obj)
	if err != nil {
		return ResultNone, err
	}

	for i := 0; i < MaxRetryTimes; i++ {
		if err := r.Client.Get(ctx, key, obj); err != nil {
			return ResultNone, err
		}

		existing := obj.DeepCopyObject()
		if err := mutate(f, key, obj); err != nil {
			return ResultNone, err
		}
		if equality.Semantic.DeepEqual(existing, obj) {
			return ResultNone, nil
		}
		if err := r.Client.Update(ctx, obj, options...); err != nil {
			if !apierrors.IsConflict(err) {
				return ResultNone, err
			}
			continue
		}
		return ResultUpdated, nil
	}

	return ResultNone, apierrors.NewTooManyRequestsError(fmt.Sprintf("Max retry: %d", MaxRetryTimes))
}

// Patch bulabula
func (r *ResourceOperations) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, options ...client.PatchOption) (Result, error) {
	if err := r.Client.Patch(ctx, obj, patch, options...); err != nil {
		return ResultNone, err
	}
	return ResultPatched, nil
}

// Delete bulabula
func (r *ResourceOperations) Delete(ctx context.Context, obj runtime.Object, options ...client.DeleteOption) (Result, error) {
	if err := r.Client.Delete(ctx, obj, options...); err != nil {
		if apierrors.IsNotFound(err) {
			return ResultNone, nil
		}
		return ResultNone, err
	}
	return ResultDeleted, nil
}

// CreateOrUpdate bulabula
func (r *ResourceOperations) CreateOrUpdate(ctx context.Context, obj runtime.Object, f MutateFn, options ...client.UpdateOption) (Result, error) {
	key, err := client.ObjectKeyFromObject(obj)
	if err != nil {
		return ResultNone, err
	}

	for i := 0; i < MaxRetryTimes; i++ {
		if err := r.Client.Get(ctx, key, obj); err != nil {
			if !apierrors.IsNotFound(err) {
				return ResultNone, err
			}
			if err := mutate(f, key, obj); err != nil {
				return ResultNone, err
			}
			if err := r.Client.Create(ctx, obj); err != nil {
				return ResultNone, err
			}
			return ResultCreated, nil
		}

		existing := obj.DeepCopyObject()
		if err := mutate(f, key, obj); err != nil {
			return ResultNone, err
		}
		if equality.Semantic.DeepEqual(existing, obj) {
			return ResultNone, nil
		}
		if err := r.Client.Update(ctx, obj, options...); err != nil {
			if !apierrors.IsConflict(err) {
				return ResultNone, err
			}
			continue
		}
		return ResultUpdated, nil
	}

	return ResultNone, apierrors.NewTooManyRequestsError(fmt.Sprintf("Max retry: %d", MaxRetryTimes))
}

// GetAndDelete bulabula
func (r *ResourceOperations) GetAndDelete(ctx context.Context, key types.NamespacedName, obj runtime.Object, options ...client.DeleteOption) (Result, error) {
	if err := r.Client.Get(ctx, key, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return ResultNone, nil
		}
		return ResultNone, err
	}
	return r.Delete(ctx, obj, options...)
}

// ContainsFinalizer bulabula
func (r *ResourceOperations) ContainsFinalizer(obj Object, finalizer string) bool {
	return controllerutil.ContainsFinalizer(obj, finalizer)
}

// AddFinalizer bulabula
func (r *ResourceOperations) AddFinalizer(ctx context.Context, obj Object, finalizer string) (Result, error) {
	return r.Update(ctx, obj, func() error {
		controllerutil.AddFinalizer(obj, finalizer)
		return nil
	})
}

// RemoveFinalizer bulabula
func (r *ResourceOperations) RemoveFinalizer(ctx context.Context, obj Object, finalizer string) (Result, error) {
	return r.Update(ctx, obj, func() error {
		controllerutil.RemoveFinalizer(obj, finalizer)
		return nil
	})
}

// ApplyDefault bulabula
func (r *ResourceOperations) ApplyDefault(ctx context.Context, obj Object) (Result, error) {
	return r.Update(ctx, obj, func() error {
		obj.Default()
		return nil
	})
}

// ApplyDefaultAll bulabula
func (r *ResourceOperations) ApplyDefaultAll(ctx context.Context, obj Object) (Result, error) {
	if result, err := r.ApplyDefault(ctx, obj); err != nil {
		return result, err
	}
	if result, err := r.Status().ApplyDefault(ctx, obj); err != nil {
		return result, err
	}
	return ResultUpdated, nil
}
