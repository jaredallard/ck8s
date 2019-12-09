package computerdeployment

import (
	"context"
	"strconv"

	computercraftv1alpha1 "github.com/cswarm/ck8sd/pkg/apis/computercraft/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
)

var log = logf.Log.WithName("controller_computerdeployment")

// Add creates a new ComputerDeployment Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileComputerDeployment{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("computerdeployment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ComputerDeployment
	err = c.Watch(&source.Kind{Type: &computercraftv1alpha1.ComputerDeployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ComputerDeployment
	err = c.Watch(&source.Kind{Type: &computercraftv1alpha1.ComputerPod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &computercraftv1alpha1.ComputerDeployment{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileComputerDeployment implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileComputerDeployment{}

// ReconcileComputerDeployment reconciles a ComputerDeployment object
type ReconcileComputerDeployment struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ComputerDeployment object and makes changes based on the state read
// and what is in the ComputerDeployment.Spec
func (r *ReconcileComputerDeployment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ComputerDeployment")

	// Fetch the ComputerDeployment instance
	deployment := &computercraftv1alpha1.ComputerDeployment{}
	err := r.client.Get(context.TODO(), request.NamespacedName, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// TODO(jaredallard): support scaling down a deployment
	for i := 0; (int64(i) - 1) != deployment.Spec.Replicas; i++ {
		// Define a new Pod object
		pod := computercraftv1alpha1.ComputerPod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployment.Name + "-" + strconv.Itoa(i),
				Namespace: deployment.Namespace,
			},
			Spec: deployment.Spec.Template.Spec,
		}

		// Set ComputerDeployment instance as the owner and controller
		if err := controllerutil.SetControllerReference(deployment, &pod, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this Pod already exists
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, &computercraftv1alpha1.ComputerPod{})
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new ComputerPod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			err = r.client.Create(context.TODO(), &pod)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Pod created successfully - don't requeue
			return reconcile.Result{}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
