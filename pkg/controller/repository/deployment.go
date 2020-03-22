package repository

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	applierv1 "github.com/bigkevmcd/kube-applier-operator/pkg/apis/applier/v1alpha1"
)

const (
	appManagedBy = "app.kubernetes.io/managed-by"
	appName      = "app.kubernetes.io/name"
	appVersion   = "app.kubernetes.io/version"
	applierImage = "quay.io/bigkevmcd/kube-applier:v0.2.0"
	gitSyncImage = "k8s.gcr.io/git-sync:v3.0.1"
)

var (
	applierAppLabels = appLabels("kube-applier", "v0.2.0")
)

func ptr32(n int32) *int32 {
	return &n
}

func deploymentFromRepository(r *applierv1.Repository) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy-" + r.Name,
			Namespace: r.Namespace,
			Annotations: map[string]string{
				appManagedBy: "kube-applier-operator",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr32(1),
			Selector: labelSelector(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: applierAppLabels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "git-repo",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name:    "kube-applier",
							Image:   applierImage,
							Command: []string{"/kube-applier"},
							// TODO: fetch this path from the resource.
							Env: []corev1.EnvVar{
								envVar("REPO_PATH", "/k8s/resources/deploy"),
								envVar("LISTEN_PORT", "2020"),
							},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{ContainerPort: 2020},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "git-repo",
									MountPath: "/k8s",
								},
							},
						},
						corev1.Container{
							Name:    "git-sync",
							Command: []string{"/git-sync"},
							Image:   gitSyncImage,
							Env: []corev1.EnvVar{
								envVar("GIT_SYNC_REPO", r.Spec.RepoURL),
								envVar("GIT_SYNC_DEST", "resources"),
								envVar("GIT_SYNC_ROOT", "/git"),
							},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{ContainerPort: 2020},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "git-repo",
									MountPath: "/git",
								},
							},
						},
					},
				},
			},
		},
	}
}

func labelSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: applierAppLabels,
	}
}

func appLabels(name, version string) map[string]string {
	return map[string]string{
		appName:    name,
		appVersion: version,
	}
}
func envVar(name, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  name,
		Value: value,
	}
}
