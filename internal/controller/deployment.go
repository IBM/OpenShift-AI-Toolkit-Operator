package controller

import (
	"fmt"

	aitoolkitv1alpha1 "github.com/IBM/OpenShift-AI-Toolkit-Operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *TritonInterfaceServerReconciler) deploymentForModelServing(tis *aitoolkitv1alpha1.TritonInterfaceServer, deploymentName string) *appsv1.Deployment {
	deployForModelServing := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              deploymentName,
			Labels:            map[string]string{"app": deploymentName},
			Namespace:         tis.Namespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": deploymentName}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:            map[string]string{"app": deploymentName, "pod": deploymentName + "-pod"},
					CreationTimestamp: metav1.Now(),
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{Name: "persistent-volume", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: tis.Spec.PvcName}}}},
					Containers: []corev1.Container{
						{
							Name:         deploymentName + "-pod",
							Image:        tis.Spec.ServingImage,
							Args:         []string{"tritonserver", "--model-repository=" + tis.Spec.MountPath},
							VolumeMounts: []corev1.VolumeMount{{Name: "persistent-volume", MountPath: tis.Spec.MountPath}},
						},
					},
				},
			},
		},
	}
	//assign resources
	requests := make(corev1.ResourceList)
	if cpuRequest, err := resource.ParseQuantity(tis.Spec.PodResources.Requests.Cpu); err == nil {
		requests[corev1.ResourceCPU] = cpuRequest
	}
	if memRequest, err := resource.ParseQuantity(tis.Spec.PodResources.Requests.Memory); err == nil {
		requests[corev1.ResourceMemory] = memRequest
	}
	deployForModelServing.Spec.Template.Spec.Containers[0].Resources.Requests = requests
	limits := make(corev1.ResourceList)
	if cpuRequest, err := resource.ParseQuantity(tis.Spec.PodResources.Limits.Cpu); err == nil {
		limits[corev1.ResourceCPU] = cpuRequest
	}
	if memRequest, err := resource.ParseQuantity(tis.Spec.PodResources.Limits.Memory); err == nil {
		limits[corev1.ResourceMemory] = memRequest
	}
	deployForModelServing.Spec.Template.Spec.Containers[0].Resources.Limits = limits
	//conditon for grpc specific tls Spec
	if tis.Spec.GrpcConfig.TlsSpec.TlsSecretName != "" {
		deployForModelServing.Spec.Template.Spec.Volumes = append(deployForModelServing.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: "grpc-tls-certificates",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: tis.Spec.GrpcConfig.TlsSpec.TlsSecretName,
				},
			},
		})
		//conditon for grpc specific arguments
		deployForModelServing.Spec.Template.Spec.Containers[0].Args = append(deployForModelServing.Spec.Template.Spec.Containers[0].Args, "--grpc-use-ssl=1", "--grpc-server-cert=/mnt/tls/tls.crt", "--grpc-server-key=/mnt/tls/tls.key")
		//condition for grpc specific volumeMounts
		deployForModelServing.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployForModelServing.Spec.Template.Spec.Containers[0].VolumeMounts,
			corev1.VolumeMount{
				Name:      "grpc-tls-certificates",
				MountPath: "/mnt/tls/tls.crt",
				SubPath:   "tls.crt",
				ReadOnly:  true,
			},
			corev1.VolumeMount{
				Name:      "grpc-tls-certificates",
				MountPath: "/mnt/tls/tls.key",
				SubPath:   "tls.key",
				ReadOnly:  true,
			})
	}
	//assign servers
	enabledservers := make(map[int]bool)
	for _, server := range tis.Spec.Servers {
		//HTTP;GRPC;Metrics
		if server.Enabled {
			if server.Type == "HTTP" {
				enabledservers[0] = true
				if server.ContainerPort != 0 {
					deployForModelServing.Spec.Template.Spec.Containers[0].Ports = append(deployForModelServing.Spec.Template.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: int32(server.ContainerPort)})
					deployForModelServing.Spec.Template.Spec.Containers[0].Args = append(deployForModelServing.Spec.Template.Spec.Containers[0].Args, "--http-port="+fmt.Sprint(server.ContainerPort))
				} else {
					deployForModelServing.Spec.Template.Spec.Containers[0].Ports = append(deployForModelServing.Spec.Template.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: int32(8000)})
				}
			} else if server.Type == "GRPC" {
				enabledservers[1] = true
				if server.ContainerPort != 0 {
					deployForModelServing.Spec.Template.Spec.Containers[0].Ports = append(deployForModelServing.Spec.Template.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: int32(server.ContainerPort)})
					deployForModelServing.Spec.Template.Spec.Containers[0].Args = append(deployForModelServing.Spec.Template.Spec.Containers[0].Args, "--grpc-port="+fmt.Sprint(server.ContainerPort))
				} else {
					deployForModelServing.Spec.Template.Spec.Containers[0].Ports = append(deployForModelServing.Spec.Template.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: int32(8001)})
				}
			} else if server.Type == "Metrics" {
				enabledservers[2] = true
				if server.ContainerPort != 0 {
					deployForModelServing.Spec.Template.Spec.Containers[0].Ports = append(deployForModelServing.Spec.Template.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: int32(server.ContainerPort)})
					deployForModelServing.Spec.Template.Spec.Containers[0].Args = append(deployForModelServing.Spec.Template.Spec.Containers[0].Args, "--metrics-port="+fmt.Sprint(server.ContainerPort))
				} else {
					deployForModelServing.Spec.Template.Spec.Containers[0].Ports = append(deployForModelServing.Spec.Template.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: int32(8002)})
				}
			}
		}
	}
	deployForModelServing.Spec.Template.Spec.Containers[0].Args = DisableServers(enabledservers, deployForModelServing.Spec.Template.Spec.Containers[0].Args)
	ctrl.SetControllerReference(tis, deployForModelServing, r.Scheme)
	return deployForModelServing
}

func DisableServers(enabledservers map[int]bool, args []string) []string {
	if !enabledservers[0] {
		args = append(args, "--allow-http=false")
	}
	if !enabledservers[1] {
		args = append(args, "--allow-grpc=false")
	}
	if !enabledservers[2] {
		args = append(args, "--allow-metrics=false")
	}
	return args
}
