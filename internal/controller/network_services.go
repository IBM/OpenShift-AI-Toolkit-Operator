package controller

import (
	"maps"
	"strings"

	aitoolkitv1alpha1 "github.com/IBM/OpenShift-AI-Toolkit-Operator/api/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *TritonInterfaceServerReconciler) routeForModelServing(tis *aitoolkitv1alpha1.TritonInterfaceServer, deploymentName, routeName, svcName string, server *aitoolkitv1alpha1.Server) *routev1.Route {
	routeForModelServing := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeName,
			Labels:    map[string]string{"app": deploymentName},
			Namespace: tis.Namespace,
		},
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{TargetPort: intstr.FromInt(int(server.ContainerPort))},
			To:   routev1.RouteTargetReference{Kind: "Service", Name: svcName},
		},
	}
	ctrl.SetControllerReference(tis, routeForModelServing, r.Scheme)
	return routeForModelServing
}

func (r *TritonInterfaceServerReconciler) serviceForModelServing(tis *aitoolkitv1alpha1.TritonInterfaceServer, deploymentName, svcName string, server *aitoolkitv1alpha1.Server) *corev1.Service {
	srvForModelServing := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Labels:    map[string]string{"app": deploymentName},
			Namespace: tis.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{{Protocol: corev1.ProtocolTCP, Port: 80, TargetPort: intstr.FromInt(int(server.ContainerPort))}},
			Selector: map[string]string{"pod": deploymentName + "-pod"},
		},
	}
	ctrl.SetControllerReference(tis, srvForModelServing, r.Scheme)
	return srvForModelServing
}

func populateDefaultServerPorts(server *aitoolkitv1alpha1.Server) {
	if server.Type == "HTTP" && server.ContainerPort == 0 {
		server.ContainerPort = 8000
	} else if server.Type == "GRPC" && server.ContainerPort == 0 {
		server.ContainerPort = 8001
	} else if server.Type == "Metrics" && server.ContainerPort == 0 {
		server.ContainerPort = 8002
	}
}

func InjectRouteConfig(tis *aitoolkitv1alpha1.TritonInterfaceServer, route *routev1.Route) {
	if len(tis.Spec.RouteConfig.RouteAnnotation.Annotations) != 0 {
		if tis.Spec.RouteConfig.RouteAnnotation.ApplyTo == "All" {
			maps.Copy(route.Annotations, tis.Spec.RouteConfig.RouteAnnotation.Annotations)
		} else if tis.Spec.RouteConfig.RouteAnnotation.ApplyTo == "HTTP" && strings.Contains(route.Name, "http") {
			maps.Copy(route.Annotations, tis.Spec.RouteConfig.RouteAnnotation.Annotations)
		} else if tis.Spec.RouteConfig.RouteAnnotation.ApplyTo == "GRPC" && strings.Contains(route.Name, "grpc") {
			maps.Copy(route.Annotations, tis.Spec.RouteConfig.RouteAnnotation.Annotations)
		} else if tis.Spec.RouteConfig.RouteAnnotation.ApplyTo == "Metrics" && strings.Contains(route.Name, "metrics") {
			maps.Copy(route.Annotations, tis.Spec.RouteConfig.RouteAnnotation.Annotations)
		}
	}
}
