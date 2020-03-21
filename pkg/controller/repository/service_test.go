package repository

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestCreateService(t *testing.T) {
	service := createService("test-ns", "kube-applier", 8080, 2020, appLabels("kube-applier", "v0.2.0"))

	want := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-applier",
			Namespace: "test-ns",
			Annotations: map[string]string{
				appManagedBy: "k8s-applier",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: appLabels("kube-applier", "v0.2.0"),
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "service",
					Port:       8080,
					TargetPort: intstr.FromInt(2020),
				},
			},
		},
	}
	if diff := cmp.Diff(want, service); diff != "" {
		t.Fatalf("service diff: %s", diff)
	}
}
