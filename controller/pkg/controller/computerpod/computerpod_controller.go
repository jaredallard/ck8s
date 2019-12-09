package computerpod

import (
	"context"
	"encoding/json"
	"fmt"
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

type jsonPatcher struct{}

// Type returns the MergePatchType
func (s *jsonPatcher) Type() types.PatchType {
	return types.MergePatchType
}

// Data returns the underlying patch data
func (s *jsonPatcher) Data(obj runtime.Object) ([]byte, error) {
	return json.Marshal(obj)
}

var log = logf.Log.WithName("controller_computerpod")

// Add creates a new ComputerPod Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	// TODO(jaredallard): just implement eventsinkimpl s
	client, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Error(err, "failed to create kubernetes client")
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	return &ReconcileComputerPod{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: eventBroadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: "ck8s-controller"}),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("computerpod-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ComputerPod
	err = c.Watch(&source.Kind{Type: &computercraftv1alpha1.ComputerPod{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileComputerPod implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileComputerPod{}

// ReconcileComputerPod reconciles a ComputerPod object
type ReconcileComputerPod struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	recorder record.EventRecorder
	scheme   *runtime.Scheme
}

// markAsFailed marks a computerpod as failed to schedule
func markAsFailed(r *ReconcileComputerPod, pod *computercraftv1alpha1.ComputerPod, reason string) error {
	pod.Status.Reason = "FailedSchedule"
	pod.Status.Phase = "Pending"
	pod.Status.Message = fmt.Sprintf("computerpod failed to schedule: %v", reason)
	r.recorder.Eventf(pod, core.EventTypeWarning, "FailedSchedule", "pod failed to be scheduled: %s", reason)
	return r.client.Status().Patch(context.TODO(), pod, &jsonPatcher{})
}

// Reconcile reads that state of the cluster for a ComputerPod object and makes changes based on the state read
// and what is in the ComputerPod.Spec
func (r *ReconcileComputerPod) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	statusPatcher := jsonPatcher{}

	// Fetch the ComputerPod instance
	pod := &computercraftv1alpha1.ComputerPod{}
	err := r.client.Get(context.TODO(), request.NamespacedName, pod)
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

	// ignore already scheduled pods
	if pod.Status.AssignedComputer != "" {
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Reconciling ComputerPod")

	computers := &computercraftv1alpha1.ComputerList{}
	namespace := client.InNamespace(pod.ObjectMeta.Namespace)

	// TODO(jaredallard): keep this stored in memory
	computerpods := &computercraftv1alpha1.ComputerPodList{}
	if err := r.client.List(context.TODO(), computers, namespace); err != nil {
		reqLogger.Error(err, "failed to list computerpods")
	}

	// hash map of pods to computers
	computerMap := map[string][]computercraftv1alpha1.ComputerPod{}
	for _, k := range computerpods.Items {
		comp := k.Status.AssignedComputer

		// skip un-assigned computerpods
		// TODO(jaredallard): filter this out at req time
		if comp == "" {
			continue
		}

		if computerMap[comp] == nil {
			computerMap[comp] = make([]computercraftv1alpha1.ComputerPod, 1)
		}

		// note that this computer is running this pod, theoretically anyways
		computerMap[comp] = append(computerMap[comp], k)
	}

	// TODO(jaredallard): make this a function
	if err := r.client.List(context.TODO(), computers, namespace); err != nil {
		reqLogger.Error(err, "failed to list computers")
		if err := markAsFailed(r, pod, err.Error()); err != nil {
			reqLogger.Error(err, "failed to mark computerpod as failed schedule", "computer", pod.Name)
		}

		// Failed to list available computers, retry scheduling
		return reconcile.Result{}, err
	}

	// check if we had no candidates
	if len(computers.Items) == 0 {
		err := fmt.Errorf("no nodes exist")
		reqLogger.Error(err, "")
		if err := markAsFailed(r, pod, err.Error()); err != nil {
			reqLogger.Error(err, "failed to mark computerpod as failed schedule", "computer", pod.Name)
		}

		// No available computers, retry scheduling
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 10}, nil
	}

	assignedComp := ""
	for _, k := range computers.Items {
		// skip computers that aren't running
		// TODO(jaredallard): filter this out at req time (fieldSelector?)
		if k.Status.Phase != "Running" {
			continue
		}

		// skip computers that kubelet aren't ready on
		kubeletReady := false
		for _, cond := range k.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == core.ConditionTrue {
				kubeletReady = true
				break
			}
		}
		if !kubeletReady {
			continue
		}

		reqLogger.Info("found running computer", "computer", k.Name)

		// don't assign more than one pod to a host right now
		if computerMap[k.Name] == nil || len(computerMap[k.Name]) == 0 {
			assignedComp = k.Name
			break
		}
	}

	if assignedComp == "" {
		err := fmt.Errorf("no nodes available")
		reqLogger.Error(err, "")
		if err := markAsFailed(r, pod, err.Error()); err != nil {
			reqLogger.Error(err, "failed to mark computerpod as failed schedule", "computer", pod.Name)
		}

		// No available computers, retry scheduling
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 10}, nil
	}

	reqLogger.Info("assigned computerpod to computer", "computer", assignedComp, "computerpod", pod.Name)
	pod.Status.AssignedComputer = assignedComp
	pod.Status.Phase = "Pending"
	pod.Status.Message = fmt.Sprintf("scheduled computer pod onto %s", assignedComp)
	pod.Status.Reason = "Scheduled"
	ack := v1.NewTime(time.Now())
	pod.Status.StartTime = &ack
	if err := r.client.Status().Patch(context.TODO(), pod, &statusPatcher); err != nil {
		reqLogger.Error(err, "failed to update status of pod to signify scheduled")

		// failed to assign this computerpod, so retry later
		return reconcile.Result{}, err
	}

	r.recorder.Eventf(pod, core.EventTypeNormal, "Scheduled", "pod was assigned computer: %s", assignedComp)

	reqLogger.Info("finished scheduling pod")
	return reconcile.Result{}, nil
}
