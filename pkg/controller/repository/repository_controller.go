package repository

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	applierv1 "github.com/bigkevmcd/kube-applier-operator/pkg/apis/applier/v1alpha1"
)

var logger = logf.Log.WithName("controller_repository")

// Add creates a new Repository Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRepository{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("repository-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Repository
	err = c.Watch(&source.Kind{Type: &applierv1.Repository{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Deployments and requeue the owner Repository
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &applierv1.Repository{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRepository implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRepository{}

// ReconcileRepository reconciles a Repository object.
type ReconcileRepository struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Repository object and makes changes based on the state read
// and what is in the Repository.Spec.
func (r *ReconcileRepository) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := logger.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Repository")

	instance := &applierv1.Repository{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	deployment := deploymentFromRepository(instance)
	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	reqLogger.Info("Skip reconcile: Deploy already exists", "Deploy.Namespace", found.Namespace, "Deploy.Name", found.Name)
	return reconcile.Result{}, nil
}
