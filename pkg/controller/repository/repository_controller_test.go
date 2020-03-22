package repository

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/bigkevmcd/kube-applier-operator/pkg/apis"
	applierv1 "github.com/bigkevmcd/kube-applier-operator/pkg/apis/applier/v1alpha1"
)

var (
	testNamespace  = "test-namespace"
	repositoryName = "test-repository"
)

var _ reconcile.Reconciler = &ReconcileRepository{}

func TestRunController(t *testing.T) {
	repository := &applierv1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      repositoryName,
			Namespace: testNamespace,
		},
		Spec: applierv1.RepositorySpec{
			RepoURL:         testRepoURL,
			SyncDestination: "env-staging",
			ResourcesPath:   "deploy",
		},
	}

	logf.SetLogger(logf.ZapLogger(true))
	objs := []runtime.Object{repository}
	r := makeReconciler(repository, objs...)

	req := reconcile.Request{
		NamespacedName: namespacedName(testNamespace, repositoryName),
	}
	res, err := r.Reconcile(req)
	fatalIfError(t, err, "reconcile: (%v)", err)
	if res.Requeue {
		t.Fatal("reconcile requeued request")
	}
	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), namespacedName(testNamespace, "deploy-"+repositoryName), deployment)
	fatalIfError(t, err, "getting deployment: (%v)", err)
	// TODO: do a "field-by-field" comparison

	service := &corev1.Service{}
	err = r.client.Get(context.TODO(), namespacedName(testNamespace, "service-"+repositoryName), service)
	fatalIfError(t, err, "getting service: (%v)", err)
}

func namespacedName(ns, name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: ns,
	}
}

func makeReconciler(r *applierv1.Repository, objs ...runtime.Object) *ReconcileRepository {
	s := scheme.Scheme
	apis.AddToScheme(scheme.Scheme)

	cl := fake.NewFakeClient(objs...)
	return &ReconcileRepository{
		client: cl,
		scheme: s,
	}
}

func fatalIfError(t *testing.T, err error, format string, a ...interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf(format, a...)
	}
}
