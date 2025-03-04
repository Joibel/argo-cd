package discovery

import (
	"context"
	"testing"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestDiscover(t *testing.T) {
	apps, err := Discover(context.Background(), nil, "./testdata", "./testdata", map[string]bool{}, []string{})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"foo": "Kustomize",
		"baz": "Helm",
	}, apps)
}

func TestAppType(t *testing.T) {
	appType, err := AppType(context.Background(), nil, "./testdata/foo", "./testdata", map[string]bool{}, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "Kustomize", appType)

	appType, err = AppType(context.Background(), nil, "./testdata/baz", "./testdata", map[string]bool{}, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "Helm", appType)

	appType, err = AppType(context.Background(), nil, "./testdata", "./testdata", map[string]bool{}, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "Directory", appType)
}

func TestAppType_Disabled(t *testing.T) {
	enableManifestGeneration := map[string]bool{
		string(v1alpha1.ApplicationSourceTypeKustomize): false,
		string(v1alpha1.ApplicationSourceTypeHelm):      false,
	}
	appType, err := AppType(context.Background(), nil, "./testdata/foo", "./testdata", enableManifestGeneration, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "Directory", appType)

	appType, err = AppType(context.Background(), nil, "./testdata/baz", "./testdata", enableManifestGeneration, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "Directory", appType)

	appType, err = AppType(context.Background(), nil, "./testdata", "./testdata", enableManifestGeneration, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "Directory", appType)
}
