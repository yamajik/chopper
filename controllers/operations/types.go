package operations

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Result bulabula
type Result string

// Result Constants bulabula
var (
	ResultNone    Result = "none"
	ResultGot     Result = "got"
	ResultListed  Result = "listed"
	ResultCreated Result = "created"
	ResultUpdated Result = "updated"
	ResultPatched Result = "patched"
	ResultDeleted Result = "deleted"
)

// Constants bulabula
var (
	MaxRetryTimes = 3
)

// MutateFn is a function which mutates the existing object into it's desired state.
type MutateFn func() error

// DefaultObject bulabula
type DefaultObject interface {
	Default()
	DefaultStatus()
}

// Object bulabula
type Object interface {
	metav1.Object
	runtime.Object
	DefaultObject
}
