package computer

import (
	"context"
	"encoding/json"
	"time"

	computercraftv1alpha1 "github.com/cswarm/ck8sd/pkg/apis/computercraft/v1alpha1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_computer")

type jsonPatcher struct{}

// Type returns the MergePatchType
func (s *jsonPatcher) Type() types.PatchType {
	return types.MergePatchType
}

// Data returns the underlying patch data
func (s *jsonPatcher) Data(obj runtime.Object) ([]byte, error) {
	return json.Marshal(obj)
}

// Add creates a new Computer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) *ReconcileComputer {
	// TODO(jaredallard): just implement eventsinkimpl s
	client, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Error(err, "failed to create kubernetes client")
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	return &ReconcileComputer{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: eventBroadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: "ck8s-controller"}),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileComputer) error {
	// Create a new controller
	c, err := controller.New("computer-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Computer
	err = c.Watch(&source.Kind{Type: &computercraftv1alpha1.Computer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	go func() {
		for {
			r.DetectDeadComputers()
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

// blank assignment to verify that ReconcileComputer implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileComputer{}

// ReconcileComputer reconciles a Computer object
type ReconcileComputer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	recorder record.EventRecorder
	scheme   *runtime.Scheme
}

// find the kublet ready condition
func getReadyCondition(c []core.NodeCondition) (int, *core.NodeCondition) {
	for i, cond := range c {
		if cond.Reason == "KubeletReady" {
			return i, &cond
		}
	}

	return 0, nil
}

func (r *ReconcileComputer) DetectDeadComputers() error {
	// TODO(jaredallard): keep this stored in memory
	computers := &computercraftv1alpha1.ComputerList{}
	namespace := client.InNamespace("default")

	if err := r.client.List(context.TODO(), computers, namespace); err != nil {
		log.Error(err, "failed to list computers")
		return err
	}

	computerpods := &computercraftv1alpha1.ComputerPodList{}
	if err := r.client.List(context.TODO(), computerpods, namespace); err != nil {
		log.Error(err, "failed to list computerpods")
		return err
	}

	for _, c := range computers.Items {
		ri, ready := getReadyCondition(c.Status.Conditions)

		// TODO(jaredallard): invalidate all pods running on that node
		// for now, we're just skipping those that are invalid
		if ready.Status == core.ConditionFalse {
			continue
		}

		now := time.Now()
		killTime := v1.NewTime(now.Add(-(time.Minute * 1)))

		if ready.LastHeartbeatTime.Before(&killTime) {
			c.Status.Conditions[ri].Status = core.ConditionFalse

			log.Info("computer hasn't pinged in configured time period, assuming unready", "computer", c.Name)
			err := r.client.Status().Patch(context.TODO(), &c, &jsonPatcher{})
			if err != nil {
				log.Error(err, "failed to mark computer unready", "computer", c.Name)
				continue // TODO(jaredallard): better handling of errors here
			}

			for _, pod := range computerpods.Items {
				if pod.Status.AssignedComputer == c.Name {
					log.Info("marking computerpod as pending due to computer being unready", "pod", pod.Name)
					pod.Status.AssignedComputer = ""
					pod.Status.Phase = core.PodPending
					pod.Status.Reason = "ComputerUnready"
					pod.Status.Message = "computer went unready"
					err := r.client.Status().Patch(context.TODO(), &pod, &jsonPatcher{})
					if err != nil {
						log.Error(err, "failed to remove pod from unready node")
					}

					r.recorder.Eventf(&pod, core.EventTypeWarning, "ComputerUnready", "computer '%s' went unready", c.Name)
				}
			}
		}
	}

	return nil
}

// Reconcile reads that state of the cluster for a Computer object and makes changes based on the state read
// and what is in the Computer.Spec
func (r *ReconcileComputer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the Computer instance
	comp := &computercraftv1alpha1.Computer{}
	err := r.client.Get(context.TODO(), request.NamespacedName, comp)
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

	// TODO(jaredallard): we need to detect when we've updated this field instead of just
	// assuming that it has been
	// maybe via an admission controller?
	for i := range comp.Status.Conditions {
		comp.Status.Conditions[i].LastHeartbeatTime = v1.NewTime(time.Now())
	}

	if err := r.client.Status().Patch(context.TODO(), comp, &jsonPatcher{}); err != nil {
		reqLogger.Error(err, "failed to update computer conditions status", "computer", comp.Name)

		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
