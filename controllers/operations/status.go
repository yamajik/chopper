package operations

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StatusOperationsInterface bulabula
type StatusOperationsInterface interface {
	Update(ctx context.Context, obj runtime.Object, f MutateFn, options ...client.UpdateOption) (Result, error)
	Patch(ctx context.Context, obj runtime.Object, patch client.Patch, options ...client.PatchOption) (Result, error)

	ApplyDefault(ctx context.Context, obj Object) (Result, error)
}

// StatusOperationsGetter bulabula
type StatusOperationsGetter interface {
	Status() StatusOperationsInterface
}

// StatusOperations bulabula
type StatusOperations struct {
	client.Client
}

// NewStatusOperations bulabula
func NewStatusOperations(c client.Client) StatusOperationsInterface {
	return &StatusOperations{c}
}

// Update bulabula
func (r *StatusOperations) Update(ctx context.Context, obj runtime.Object, f MutateFn, options ...client.UpdateOption) (Result, error) {
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
		if err := r.Client.Status().Update(ctx, obj, options...); err != nil {
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
func (r *StatusOperations) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, options ...client.PatchOption) (Result, error) {
	if err := r.Client.Patch(ctx, obj, patch, options...); err != nil {
		return ResultNone, err
	}
	return ResultPatched, nil
}

// ApplyDefault bulabula
func (r *StatusOperations) ApplyDefault(ctx context.Context, obj Object) (Result, error) {
	return r.Update(ctx, obj, func() error {
		obj.DefaultStatus()
		return nil
	})
}
