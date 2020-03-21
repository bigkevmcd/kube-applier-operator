package repository

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-cmp/cmp"

	applierv1 "github.com/bigkevmcd/k8s-applier/pkg/apis/applier/v1alpha1"
)

const testRepoURL = "https://github.com/bigkevmcd/taxi.git"

func TestDeploymentFromRepository(t *testing.T) {
	r := &applierv1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy-taxi",
			Namespace: "taxi-staging",
		},
		Spec: applierv1.RepositorySpec{
			RepoURL:         testRepoURL,
			SyncDestination: "env-staging",
			ResourcesPath:   "deploy",
		},
	}

	want := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy-deploy-taxi",
			Namespace: "taxi-staging",
			Annotations: map[string]string{
				appManagedBy: "k8s-applier",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr32(1),
			Selector: labelSelector(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: appLabels("kube-applier", "v0.2.0"),
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
							Command: []string{"/kube-applier"},
							Image:   "quay.io/bigkevmcd/kube-applier:v0.2.0",
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
								envVar("GIT_SYNC_REPO", testRepoURL),
								envVar("GIT_SYNC_DEST", "resources"),
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
	deployment := deploymentFromRepository(r)

	if diff := cmp.Diff(want, deployment); diff != "" {
		t.Fatalf("deployment diff: %s", diff)
	}
}
