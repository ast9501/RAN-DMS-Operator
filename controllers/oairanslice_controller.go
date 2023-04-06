/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	helm "github.com/ast9501/helm-clientgo/pkg"
	//helm "github.com/ast9501/ran-dms-operator/pkg/helm/"

	ranslicev1alpha1 "github.com/ast9501/ran-dms-operator/api/v1alpha1"
)

var logger = log.Log.WithName("controller_ranslice")

const finalizerName string = "ranslice.finalizer.win.nycu"

// OAIRanSliceReconciler reconciles a OAIRanSlice object
type OAIRanSliceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// State of OAIRanSlice
const (
	StateNull     string = ""
	StateCreating string = "Creating"
	StateRunning  string = "Running"
)

var sliceIdx int = 1
var ipPoolNetworkID24 string = "192.168.3."

// TODO: check "/var/lib/cni/networks/<bridge-network-name>/*" to get used ips
var controlIpPoolHostID int = 150
var nguIpPoolHostID int = 50
var oaiHelmCommonPackageName string = "win-oai"
var oaiCuPackageName string = "cu-cp"
var oaiSliceMap map[string]int = make(map[string]int)

//var oaiDuPackageName string = "du"

//+kubebuilder:rbac:groups=ranslice.winlab.nycu,resources=oairanslice,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ranslice.winlab.nycu,resources=oairanslice/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ranslice.winlab.nycu,resources=oairanslice/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OAIRanSlice object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *OAIRanSliceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	logger.Info("Reconciling OAIRANSlice", "Request.Namespace", req.Namespace, "Request.Name", req.Name)

	// TODO(user): your logic here

	// Fetch the ranSlice instance
	instance := &ranslicev1alpha1.OAIRanSlice{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
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

	// Check if OAIRanSlice object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being delete
		if !containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.Client.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			// The finializer is present
			// Delete external Helm resource
			cuReleaseName := "oai-cu-slice" + strconv.Itoa(oaiSliceMap[instance.Name])
			//logger.Info("Delete cuRelease: ", cuReleaseName) //oai-cu-slice0
			error := helm.UninstallHelmChart(instance.Namespace, cuReleaseName)
			logger.Error(error, "Failed to uninstall charts", "ReleaseName", cuReleaseName)
		}

		// Stop reconcliiation as the object is being deleted
		return reconcile.Result{}, nil
	}

	// Check OAIRanSlice state
	if instance.Status.State == StateRunning || instance.Status.State == StateCreating {
		logger.Info("OAIRanSlice is activate")
		return reconcile.Result{}, nil
	} else if instance.Status.State != StateNull {
		err := fmt.Errorf("unknown OAIRanSlice.Status.State %s", instance.Status.State)
		return reconcile.Result{}, err
	}

	// Update OAIRanSlice.Status.State to Creating
	instance.Status.State = StateCreating
	err = r.Client.Status().Update(context.TODO(), instance)
	if err != nil {
		logger.Error(err, "Failed to update OAIRanSlice status")
		return reconcile.Result{}, err
	}

	// TODO: update helm repo

	// TODO: Check if common NF already exist (for example, RT-RIC)
	// TODO: Check if CU, as RAN common NF, already exist (if not exist, create new)

	// create cu
	// init new helm values
	var vals string = ""

	// add sliceIdx to values
	vals += "sliceIdx=" + strconv.Itoa(sliceIdx)
	logger.Info("add sliceIdx:", "vals:", vals)

	// add supportedSnssaiList
	for i, e := range instance.Spec.SnssaiList {
		snssai := ",global.snssaiLists[" + strconv.Itoa(i) + "].sst=" + fmt.Sprint(e.Sst) + ",global.snssaiLists[" + strconv.Itoa(i) + "].sd=" + e.Sd
		vals += snssai
	}
	logger.Info("add snssaiList:", "vals:", vals)

	// assign cu ip addrs
	cucpAddr := newControlIp()
	vals += ",cp.addr=" + cucpAddr
	logger.Info("set cp addr:", "ip", cucpAddr)
	cuupAddr := newControlIp()
	vals += ",up.addr=" + cuupAddr
	logger.Info("set up addr:", "ip", cuupAddr)
	nguAddr := newNguIp()
	vals += ",ngu.addr=" + nguAddr
	logger.Info("set ngu addr:", "ip", nguAddr)

	args := wrapHelmVal(vals)
	logger.Info("final args: ", "args", args["set"])
	err = helm.InstallChart("oai-cu-slice"+strconv.Itoa(sliceIdx), instance.Namespace, oaiHelmCommonPackageName, oaiCuPackageName, instance.Spec.CuPackageVersion, args)
	if err != nil {
		logger.Error(err, "Failed to install chart", "ReleaseName", "oai-cu-slice"+strconv.Itoa(sliceIdx))
	}

	logger.Info("Successfully create oai-cu-slice, ", "SliceID", sliceIdx, "S-NSSAIList", instance.Spec.SnssaiList)

	// TODO: add du slice

	// update OAISlice status
	instance.Status.CUCPAddr = cucpAddr
	instance.Status.CUUPAddr = cuupAddr
	instance.Status.NGUAddr = nguAddr
	instance.Status.State = StateRunning

	// update sliceIdx
	oaiSliceMap[instance.Name] = sliceIdx
	logger.Info("Append oaiSliceMap, ", "instance.Name", instance.Name)
	sliceIdx++
	logger.Info("Update sliceIdx, ", "sliceIdx", sliceIdx)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OAIRanSliceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ranslicev1alpha1.OAIRanSlice{}).
		Complete(r)
}

// Helper functions to check and remove string from a slice of strings.
// See https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/src/cronjob-tutorial/testdata/finalizer_example.go

// containsString checks if the given slice of string contains the target string
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// new control ip, allocate ip range between .150 to .254
func newControlIp() (newIp string) {
	newIp = ipPoolNetworkID24 + strconv.Itoa(controlIpPoolHostID)
	if controlIpPoolHostID < 254 {
		controlIpPoolHostID++
	} else {
		controlIpPoolHostID = 150
	}
	return
}

// new ngu interface ip, allocate ip range between .50 to .99
func newNguIp() (newIp string) {
	newIp = ipPoolNetworkID24 + strconv.Itoa(nguIpPoolHostID)
	if nguIpPoolHostID < 99 {
		nguIpPoolHostID++
	} else {
		nguIpPoolHostID = 50
	}
	return
}

// encap helm cli cmd
func wrapHelmVal(vals string) map[string]string {
	args := make(map[string]string)
	args["set"] = vals
	return args
}
