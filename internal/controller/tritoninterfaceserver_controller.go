/*
Copyright 2024.

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

package controller

import (
	"context"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aitoolkitv1alpha1 "github.com/IBM/OpenShift-AI-Toolkit-Operator/api/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
)

// TritonInterfaceServerReconciler reconciles a TritonInterfaceServer object
type TritonInterfaceServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ai-toolkit.ibm.com,resources=tritoninterfaceservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ai-toolkit.ibm.com,resources=tritoninterfaceservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ai-toolkit.ibm.com,resources=tritoninterfaceservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes/custom-host,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TritonInterfaceServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *TritonInterfaceServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log.Log.Info("Inside TritonInterfaceServerReconciler", "Request", req)
	tis := &aitoolkitv1alpha1.TritonInterfaceServer{}
	err := r.Get(ctx, req.NamespacedName, tis)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Log.Info("TritonInterfaceServer not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Log.Error(err, "Failed to get TritonInterfaceServer")
		return ctrl.Result{}, err
	}
	//
	log.Log.Info("Check", "Spec", tis.Spec)
	//
	deploymentName := string("triton-server-"+tis.UID[:8]) + "-" + tis.Spec.PvcName
	//check Deployment
	deploymentResource := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: tis.Namespace}, deploymentResource)
	if err != nil && errors.IsNotFound(err) {
		deploy := r.deploymentForModelServing(tis, deploymentName)
		log.Log.Info("Creating a new Deployment", "deployment.Namespace", tis.Namespace, "deployment.Name", deploy.Name, "TritonInterfaceServer", tis.Name)
		err = r.Create(ctx, deploy)
		if err != nil {
			log.Log.Error(err, "Failed to create new Deployment", "deployment.Namespace", deploy.Namespace, "deployment.Name", deploy.Name, "TritonInterfaceServer", tis.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Log.Error(err, "Failed to get Deployment", "deployment.Namespace", tis.Namespace, "deployment.Name", deploymentName, "TritonInterfaceServer", tis.Name)
		return ctrl.Result{}, err
	}
	//check services and routes
	for _, server := range tis.Spec.Servers {
		populateDefaultServerPorts(&server)
		//check services
		svcName := strings.ReplaceAll(strings.ToLower(server.Type+"-service-"+deploymentName), "_", "-")
		serviceResource := &corev1.Service{}
		err = r.Get(ctx, types.NamespacedName{Name: svcName, Namespace: tis.Namespace}, serviceResource)
		if err != nil && errors.IsNotFound(err) {
			svc := r.serviceForModelServing(tis, deploymentName, svcName, &server)
			log.Log.Info("Creating a new Service", "Service.Namespace", tis.Namespace, "Service.Name", svc.Name, "TritonInterfaceServer", tis.Name)
			err = r.Create(ctx, svc)
			if err != nil {
				log.Log.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name, "TritonInterfaceServer", tis.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Log.Error(err, "Failed to get Service", "Service.Namespace", tis.Namespace, "Service.Name", svcName, "TritonInterfaceServer", tis.Name)
			return ctrl.Result{}, err
		}
		//check route
		routeName := strings.ReplaceAll(strings.ToLower(server.Type+"-route-"+deploymentName), "_", "-")
		routeResource := &routev1.Route{}
		err = r.Get(ctx, types.NamespacedName{Name: routeName, Namespace: tis.Namespace}, routeResource)
		if err != nil && errors.IsNotFound(err) {
			// Define a new deployment
			if strings.Contains(routeName, "grpc") {
				log.Log.Info("Configuring TLS Passthrough for gRPC route", "Route.Name", routeName)
				tlsConfig := &routev1.TLSConfig{
					Termination: routev1.TLSTerminationPassthrough,
				}
				// Define the route with TLS passthrough
				route := r.routeForModelServing(tis, deploymentName, routeName, svcName, &server)
				route.Spec.TLS = tlsConfig
				log.Log.Info("Creating a new Route with TLS Passthrough", "Route.Namespace", tis.Namespace, "Route.Name", route.Name, "TritonInterfaceServer", tis.Name)
				err = r.Create(ctx, route)
				if err != nil {
					log.Log.Error(err, "Failed to create new Route with TLS Passthrough", "Route.Namespace", route.Namespace, "Route.Name", route.Name, "TritonInterfaceServer", tis.Name)
					return ctrl.Result{}, err
				}
			} else {
				route := r.routeForModelServing(tis, deploymentName, routeName, svcName, &server)
				log.Log.Info("Creating a new Route", "Route.Namespace", tis.Namespace, "Route.Name", route.Name, "TritonInterfaceServer", tis.Name)
				err = r.Create(ctx, route)
				if err != nil {
					log.Log.Error(err, "Failed to create new Route", "Route.Namespace", route.Namespace, "Route.Name", route.Name, "TritonInterfaceServer", tis.Name)
					return ctrl.Result{}, err
				}
				// Deployment created successfully - return and requeue
				return ctrl.Result{Requeue: true}, nil
			}
		} else if err != nil {
			log.Log.Error(err, "Failed to get Route", "Route.Namespace", tis.Namespace, "Route.Name", routeName, "TritonInterfaceServer", tis.Name)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TritonInterfaceServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aitoolkitv1alpha1.TritonInterfaceServer{}).
		Owns(&appsv1.Deployment{}).Owns(&corev1.Service{}).Owns(&routev1.Route{}).
		Complete(r)
}
