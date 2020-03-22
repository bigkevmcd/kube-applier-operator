package repository

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func createService(ns, name string, port, targetPort int32, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				appManagedBy: "kube-applier-operator",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: appLabels("kube-applier", "v0.2.0"),
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "service",
					Port:       port,
					TargetPort: intstr.FromInt(int(targetPort)),
				},
			},
		},
	}
}
