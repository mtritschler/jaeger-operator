package deployment

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"

	"github.com/jaegertracing/jaeger-operator/pkg/apis/io/v1alpha1"
)

func init() {
	viper.SetDefault("jaeger-version", "1.6")
	viper.SetDefault("jaeger-all-in-one-image", "jaegertracing/all-in-one")
}

func TestDefaultAllInOneImage(t *testing.T) {
	viper.Set("jaeger-all-in-one-image", "org/custom-all-in-one-image")
	viper.Set("jaeger-version", "123")
	defer viper.Reset()

	d := NewAllInOne(v1alpha1.NewJaeger("TestAllInOneDefaultImage")).Get()

	assert.Len(t, d.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "org/custom-all-in-one-image:123", d.Spec.Template.Spec.Containers[0].Image)

	envvars := []v1.EnvVar{
		v1.EnvVar{
			Name:  "SPAN_STORAGE_TYPE",
			Value: "",
		},
		v1.EnvVar{
			Name:  "COLLECTOR_ZIPKIN_HTTP_PORT",
			Value: "9411",
		},
	}
	assert.Equal(t, envvars, d.Spec.Template.Spec.Containers[0].Env)
}

func TestAllInOneAnnotations(t *testing.T) {
	jaeger := v1alpha1.NewJaeger("TestAllInOneAnnotations")
	jaeger.Spec.AllInOne.Annotations = map[string]string{
		"hello":                "world",
		"prometheus.io/scrape": "false", // Override implicit value
	}

	allinone := NewAllInOne(jaeger)
	dep := allinone.Get()

	assert.Equal(t, "false", dep.Spec.Template.Annotations["sidecar.istio.io/inject"])
	assert.Equal(t, "world", dep.Spec.Template.Annotations["hello"])
	assert.Equal(t, "false", dep.Spec.Template.Annotations["prometheus.io/scrape"])
}

func TestAllInOneHasOwner(t *testing.T) {
	name := "TestAllInOneHasOwner"
	a := NewAllInOne(v1alpha1.NewJaeger(name))
	assert.Equal(t, name, a.Get().ObjectMeta.Name)
}

func TestAllInOneNumberOfServices(t *testing.T) {
	name := "TestNumberOfServices"
	services := NewAllInOne(v1alpha1.NewJaeger(name)).Services()
	assert.Len(t, services, 3) // collector, query, agent

	for _, svc := range services {
		owners := svc.ObjectMeta.OwnerReferences
		assert.Equal(t, name, owners[0].Name)
	}
}
