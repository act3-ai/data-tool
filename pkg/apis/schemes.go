// Package apis implements shared functionality intended for api use, including configuration schema manipulation.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/act3-ai/data-tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// NewScheme creates the scheme for the act3-pt config files.
func NewScheme() *runtime.Scheme {
	// schemeBuilder is used to add go types to the GroupVersionKind scheme
	schemeBuilder := runtime.NewSchemeBuilder(
		v1alpha1.AddToScheme,
	)

	// addToScheme adds the types in this group-version to the given scheme.
	addToScheme := schemeBuilder.AddToScheme

	scheme := runtime.NewScheme()
	utilruntime.Must(addToScheme(scheme))

	return scheme
}
