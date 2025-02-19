# OpenShift-AI-Toolkit-Operator
This operator provides Deployment Server for various models available on IBM Z

## Operator Based on
[IBM Z Nvidia triton Repo](https://github.com/IBM/ibmz-accelerated-for-nvidia-triton-inference-server?tab=readme-ov-file#models-repository)

## PREREQUISITES
- IBM Cloud Account
- Openshift Cluster on IBM Z
- Openshift Client CLI
 
## Installation

### Step 1: Setup ICR Auth credientials in openshift-config

- Define Enviroment

  ```sh
  (base) shivanggoswami@Shivangs-MacBook-Pro operator-setup % cat env-template.sh 
  export REGISTRY_URL="icr.io" #The Registry URL. Default: icr.io
  export EMAIL="person@ibm.com" #IBM Cloud Userid. Format: user@ibm.com
  export API_KEY="api-key" #IBM Cloud API key
  ```

  Modify the Following file and load the environment variables in the system

  After this, First login into your openshift cluster and execute this script

  ```sh
  (base) shivanggoswami@Shivangs-MacBook-Pro operator-setup % source env.sh # this is derived from env-templates 
  (base) shivanggoswami@Shivangs-MacBook-Pro operator-setup % ./populateICRconfig.sh 
  secret/icr-registry created
  secret "icr-registry" deleted
  secret/pull-secret data updated
  ```

  The following command can be used to assert the script worked properly
  ```sh
  (base) shivanggoswami@Shivangs-MacBook-Pro operator-setup % oc get secret pull-secret -n openshift-config -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d
  ...
   "icr.io": {
      "username": "iamapikey",
      "password": <ibm-api-key>,
      "email": <ibm-user-id>,
      "auth": <username/password in base64 encoding>
    }
   }
  }
  ```
### Step 2: Apply Catalog Source

- Apply Catalog Source
  ```sh
  (base) shivanggoswami@Shivangs-MacBook-Pro operator-setup % oc apply -f catalog-source.yaml
  catalogsource.operators.coreos.com/openshift-ai-toolkit-operator-catalog created
  ```

  If everything goes through fine, operator will be available in operator hub and can be installed from there
  ![add5d509-cd70-41a6-aad5-f2b77774dab8](https://github.com/user-attachments/assets/5718a581-080c-496f-b68f-0c349e829696)

## Using this Operator

### Step 1A: Creating pvc and injecting Model data inside them (Required)

As an example, there is a model directory present in the repo with two pre-populated data modules: credit card fraud detection(Snapml) and densenet(onnx)
```sh
(base) shivanggoswami@Shivangs-MacBook-Pro operator-examples % tree models         
models
├── cc_fraud_detect_model
│   ├── 1
│   │   └── model.pmml
│   └── config.pbtxt
└── densenet_onnx
    ├── 1
    │   └── model.so
    └── config.pbtxt

5 directories, 4 files
```
This Directory structure and model formats can be referred from the parent repo.

First create a openshift namespace, then populate data values in this script
```sh
(base) shivanggoswami@Shivangs-MacBook-Pro operator-examples % cat create-sync-pvc.sh 
...
# Set your variables
LOCAL_DIR="./models"    # Local directory to copy data from
NAMESPACE="test"        # Kubernetes namespace
PVC_NAME="triton-pvc"   # PVC name
POD_NAME="alpine-pvc-pod" # Pod name
MOUNT_PATH="/mnt/data"  # PVC mount path inside the pod
CONTAINER_NAME="alpine" # Container name
PVC_STORAGE="10Gi"      # PVC storage size
ALPINE_IMAGE="alpinelinux/rsyncd" # Alpine image to use for the pod
CLEAN=true             # Set to true to clean (delete) existing PVC and Pod before execution
...
```

The following script use oc rsync command so local directory can sync delta changes with pvc if the pvc already exists
| Variable Name   | Sample Value           | Explanation                                                   |
|------------------|------------------------|---------------------------------------------------------------|
| `LOCAL_DIR`     | `./models`            | Local directory path containing the data to copy.             |
| `NAMESPACE`     | `test`                | The Kubernetes namespace to operate in.                      |
| `PVC_NAME`      | `triton-pvc`          | Name of the Persistent Volume Claim to create or use.         |
| `POD_NAME`      | `alpine-pvc-pod`      | Name of the pod to be created or managed.                     |
| `MOUNT_PATH`    | `/mnt/data`           | Path where the PVC will be mounted inside the pod.            |
| `CONTAINER_NAME`| `alpine`              | Name of the container within the pod.                         |
| `PVC_STORAGE`   | `10Gi`                | Storage size requested for the PVC.                           |
| `ALPINE_IMAGE`  | `alpinelinux/rsyncd`  | Docker image to use for the pod's container.                  |
| `CLEAN`         | `true`                | Indicates whether to delete existing PVC and pod before execution. |

Attaching sample Output for Reference:
<details> <summary>Click to expand: Shell Output</summary>
 
 ```sh
 [root@t313lp68 operator-examples]# ./create-sync-pvc.sh 
 Creating PVC triton-pvc...
 persistentvolumeclaim/triton-pvc created
 PVC is already bound, no need to wait.
 Creating pod alpine-pvc-pod with Alpine image and PVC mount...
 Warning: would violate PodSecurity "restricted:v1.24": allowPrivilegeEscalation != false (container "alpine" must set securityContext.allowPrivilegeEscalation=false), unrestricted capabilities (container "alpine" must set securityContext.capabilities.drop=["ALL"]), runAsNonRoot != true (pod or container "alpine" must set securityContext.runAsNonRoot=true), seccompProfile (pod or container "alpine" must set securityContext.seccompProfile.type to "RuntimeDefault" or "Localhost")
 pod/alpine-pvc-pod created
 Waiting for pod alpine-pvc-pod to start...
 Waiting for pod alpine-pvc-pod to be ready...
 pod/alpine-pvc-pod condition met
 Copying data from ./models to /mnt/data inside pod alpine-pvc-pod...
 sending incremental file list
 ./
 cc_fraud_detect_model/
 cc_fraud_detect_model/config.pbtxt
             459 100%    0.00kB/s    0:00:00 (xfr#1, to-chk=5/9)
 cc_fraud_detect_model/1/
 cc_fraud_detect_model/1/model.pmml
       2,375,804 100%   75.52MB/s    0:00:00 (xfr#2, to-chk=3/9)
 densenet_onnx/
 densenet_onnx/config.pbtxt
             246 100%    8.01kB/s    0:00:00 (xfr#3, to-chk=2/9)
 densenet_onnx/1/
 densenet_onnx/1/model.so
      33,201,936 100%   19.65MB/s    0:00:01 (xfr#4, to-chk=0/9)
 
 sent 30,371,036 bytes  received 127 bytes  20,247,442.00 bytes/sec
 total size is 35,578,445  speedup is 1.17
 Data successfully copied to PVC.
 Script execution completed in 6 seconds.
 ```
</details>

### Step 1B: Configuring TLS Certificates for GRPC Server (Optional for gRPC Server usage)

gRPC Works on HTTP/2 Protocol. Same needs to be enabled in Openshift ingress controller.

Syntax:
``` sh
oc -n openshift-ingress-operator annotate ingresscontrollers/<ingresscontroller_name> ingress.operator.openshift.io/default-enable-http2=true
```

To enable HTTP/2 on the entire cluster, use the following command.

``` sh
oc annotate ingresses.config/cluster ingress.operator.openshift.io/default-enable-http2=true
```

Create a TLS secret consisting of TLS certificate and TLS private key using the following command.This Configuration will be passed onto the gRPC Server.

Syntax:
```sh
# Create TLS certificate using the openssl command. Fill the details requested.
openssl req -newkey rsa:4096 -nodes -keyout <path_to_store_tls_key> -x509 -days 365 -out <path_to_store_tls_cert>
```

Sample:
```sh
mkdir tls
openssl req -newkey rsa:4096 -nodes -keyout tls/tls.key -x509 -days 365 -out tls/tls.crt
```

Inject these configuration in Openshift secret:
```sh
oc create secret tls grpc-tls-secret --cert=tls/tls.crt --key=tls/tls.key
```

This information needs to populated in the CRD yaml file 

### Step 2: Creating the Custom CRDs

Assuming that operator was installed successfully and the custom crds can be applied now

#### Triton Server Interface CRD
### TritonInterfaceServerSpec Variables

| **Variable Name**    | **Sample Value**                                                                                     | **Explanation**                                                                                   |
|-----------------------|-----------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------|
| `pvcName`            | `"triton-pvc"`                                                                                     | Name of the Persistent Volume Claim (PVC) to be used.                                             |
| `mountPath`          | `"/mnt/data"`                                                                                      | Path where the PVC will be mounted in the container.                                              |
| `servingImage`       | `"icr.io/ibmz/ibmz-accelerated-for-nvidia-triton-inference-server@sha256:2cedd535805c316..."`       | The Docker image used for the Triton inference server.                                            |
| `servers`            | `[{"type": "HTTP", "enabled": true, "containerPort": 8000}]`                                       | List of server configurations including type, whether enabled, and container port.                |
| `podResources`       | `{"limits": {"cpu": "2", "memory": "2Gi"}, "requests": {"cpu": "1", "memory": "1Gi"}}`             | Resource requests and limits for the pod (CPU and memory).                                        |

---

### `Server` Variables

| **Variable Name**     | **Sample Value** | **Explanation**                                                                                     |
|------------------------|------------------|-----------------------------------------------------------------------------------------------------|
| `type`                | `"HTTP"`         | The type of server. Valid values are `HTTP`, `GRPC`, or `Metrics`.                                  |
| `enabled`             | `true`           | Whether the server type is enabled.                                                                |
| `containerPort`       | `8000`           | Port exposed by the container. Must be between 0 and 65535.                                         |

---

### `Resource` and `PodResource` Variables

| **Variable Name**     | **Sample Value**       | **Explanation**                                                                                     |
|------------------------|------------------------|-----------------------------------------------------------------------------------------------------|
| `limits.cpu`          | `"2"`                 | Maximum number of CPU cores allocated to the pod.                                                   |
| `limits.memory`       | `"2Gi"`               | Maximum memory allocated to the pod.                                                               |
| `requests.cpu`        | `"1"`                 | Minimum guaranteed number of CPU cores allocated to the pod.                                        |
| `requests.memory`     | `"1Gi"`               | Minimum guaranteed memory allocated to the pod.                                                    |

---

### `TlsSpec` Variables

| **Variable Name**     | **Sample Value**       | **Explanation**                                                                                     |
|------------------------|------------------------|-----------------------------------------------------------------------------------------------------|
| `tlsSecretName`          | `"grpc-tls-secret"`                 | The OpenShift Secret Name where TLS configurations are stored.                       |                         

---

There is a sample crd within the repo as well

```sh
(base) shivanggoswami@Shivangs-MacBook-Pro samples % pwd
/Users/shivanggoswami/Documents/toolkit-new/config/samples
(base) shivanggoswami@Shivangs-MacBook-Pro samples % cat ai-toolkit_v1alpha1_tritoninterfaceserver.yaml 
apiVersion: ai-toolkit.ibm.com/v1alpha1
kind: TritonInterfaceServer
metadata:
  labels:
    app.kubernetes.io/name: openshift-ai-toolkit
    app.kubernetes.io/managed-by: kustomize
  name: tritoninterfaceserver-sample
spec:
  pvcName: triton-pvc
  mountPath: "/mount"
  servers:
    - type: HTTP
      enabled: true
    - type: Metrics
      enabled: true
    - type: GRPC
      enabled: true
  grpcConfig:
    tlsSpec:
      tlsSecretName: grpc-tls-secret
  podResources:
    limits:
      cpu: 1000m
      memory: 2Gi
    requests:
      cpu: 100m
      memory: 200Mi
```

Once the CRD is applied to a particular namespace, deployments, services and routes will be created for the same.

## Using the Endpoints

In a test cluster, the following resources were created in the order

- Injected ICR config to global secret using the script mentioned above
- Created namespace "triton-ns"
- Used Sample models provided with the documentation and create pvc using the script within the same namespace
- installed the operator
- Applied the sample crd (via UI or server-side in cli) within the same namespace
- Check the resources created within the namespace

```sh
[root@t313lp68 operator-examples]# oc get all -n triton-ns
Warning: apps.openshift.io/v1 DeploymentConfig is deprecated in v4.14+, unavailable in v4.10000+
NAME                                                           READY   STATUS    RESTARTS      AGE
pod/alpine-pvc-pod                                             1/1     Running   4 (37m ago)   4h37m
pod/openshift-ai-toolkit-controller-manager-5f59859f7b-dbpnv   1/1     Running   0             30m
pod/triton-server-658db168-triton-pvc-6bdb495998-26bqd         1/1     Running   0             113s

NAME                                                              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/grpc-service-triton-server-658db168-triton-pvc            ClusterIP   172.30.123.67   <none>        80/TCP     113s
service/http-service-triton-server-658db168-triton-pvc            ClusterIP   172.30.255.51   <none>        80/TCP     113s
service/metrics-service-triton-server-658db168-triton-pvc         ClusterIP   172.30.142.91   <none>        80/TCP     113s
service/openshift-ai-toolkit-controller-manager-metrics-service   ClusterIP   172.30.84.169   <none>        8443/TCP   30m

NAME                                                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/openshift-ai-toolkit-controller-manager   1/1     1            1           30m
deployment.apps/triton-server-658db168-triton-pvc         1/1     1            1           114s

NAME                                                                 DESIRED   CURRENT   READY   AGE
replicaset.apps/openshift-ai-toolkit-controller-manager-5f59859f7b   1         1         1       30m
replicaset.apps/triton-server-658db168-triton-pvc-6bdb495998         1         1         1       114s

NAME                                                                       HOST/PORT                                                                               PATH   SERVICES                                            PORT   TERMINATION   WILDCARD
route.route.openshift.io/grpc-route-triton-server-658db168-triton-pvc      grpc-route-triton-server-658db168-triton-pvc-ai-toolkit.apps.a314lp45ocp.lnxne.boe             grpc-service-triton-server-658db168-triton-pvc      8001   passthrough   None
route.route.openshift.io/http-route-triton-server-658db168-triton-pvc      http-route-triton-server-658db168-triton-pvc-ai-toolkit.apps.a314lp45ocp.lnxne.boe             http-service-triton-server-658db168-triton-pvc      8000                 None
route.route.openshift.io/metrics-route-triton-server-658db168-triton-pvc   metrics-route-triton-server-658db168-triton-pvc-ai-toolkit.apps.a314lp45ocp.lnxne.boe          metrics-service-triton-server-658db168-triton-pvc   8002                 None
```

Via this example we have created two route one for http server and one for metrics server

Sample Http Request
<details>
  <summary>Click to expand the shell command</summary>

```sh
[root@t313lp68 operator-examples]# curl -X POST http://http-route-triton-server-624c8ff1-triton-pvc-triton-ns.apps.t313lp68ocp.lnxne.boe/v2/repository/index | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   119  100   119    0     0  39666      0 --:--:-- --:--:-- --:--:-- 59500
[
  {
    "name": "cc_fraud_detect_model",
    "version": "1",
    "state": "READY"
  },
  {
    "name": "densenet_onnx",
    "version": "1",
    "state": "READY"
  }
]
```
</details>

Sample Metrics Request
<details>
  <summary>Click to expand the shell command</summary>

```sh
[root@t313lp68 operator-examples]# curl metrics-route-triton-server-624c8ff1-triton-pvc-triton-ns.apps.t313lp68ocp.lnxne.boe/metrics
# HELP nv_inference_request_success Number of successful inference requests, all batch sizes
# TYPE nv_inference_request_success counter
nv_inference_request_success{model="densenet_onnx",version="1"} 0
nv_inference_request_success{model="cc_fraud_detect_model",version="1"} 0
# HELP nv_inference_request_failure Number of failed inference requests, all batch sizes
# TYPE nv_inference_request_failure counter
nv_inference_request_failure{model="densenet_onnx",reason="OTHER",version="1"} 0
nv_inference_request_failure{model="densenet_onnx",reason="BACKEND",version="1"} 0
nv_inference_request_failure{model="densenet_onnx",reason="CANCELED",version="1"} 0
nv_inference_request_failure{model="cc_fraud_detect_model",reason="OTHER",version="1"} 0
nv_inference_request_failure{model="cc_fraud_detect_model",reason="BACKEND",version="1"} 0
nv_inference_request_failure{model="cc_fraud_detect_model",reason="CANCELED",version="1"} 0
nv_inference_request_failure{model="densenet_onnx",reason="REJECTED",version="1"} 0
nv_inference_request_failure{model="cc_fraud_detect_model",reason="REJECTED",version="1"} 0
# HELP nv_inference_count Number of inferences performed (does not include cached requests)
# TYPE nv_inference_count counter
nv_inference_count{model="densenet_onnx",version="1"} 0
nv_inference_count{model="cc_fraud_detect_model",version="1"} 0
# HELP nv_inference_exec_count Number of model executions performed (does not include cached requests)
# TYPE nv_inference_exec_count counter
nv_inference_exec_count{model="densenet_onnx",version="1"} 0
nv_inference_exec_count{model="cc_fraud_detect_model",version="1"} 0
# HELP nv_inference_request_duration_us Cumulative inference request duration in microseconds (includes cached requests)
# TYPE nv_inference_request_duration_us counter
nv_inference_request_duration_us{model="densenet_onnx",version="1"} 0
nv_inference_request_duration_us{model="cc_fraud_detect_model",version="1"} 0
# HELP nv_inference_queue_duration_us Cumulative inference queuing duration in microseconds (includes cached requests)
# TYPE nv_inference_queue_duration_us counter
nv_inference_queue_duration_us{model="densenet_onnx",version="1"} 0
nv_inference_queue_duration_us{model="cc_fraud_detect_model",version="1"} 0
# HELP nv_inference_compute_input_duration_us Cumulative compute input duration in microseconds (does not include cached requests)
# TYPE nv_inference_compute_input_duration_us counter
nv_inference_compute_input_duration_us{model="densenet_onnx",version="1"} 0
nv_inference_compute_input_duration_us{model="cc_fraud_detect_model",version="1"} 0
# HELP nv_inference_compute_infer_duration_us Cumulative compute inference duration in microseconds (does not include cached requests)
# TYPE nv_inference_compute_infer_duration_us counter
nv_inference_compute_infer_duration_us{model="densenet_onnx",version="1"} 0
nv_inference_compute_infer_duration_us{model="cc_fraud_detect_model",version="1"} 0
# HELP nv_inference_compute_output_duration_us Cumulative inference compute output duration in microseconds (does not include cached requests)
# TYPE nv_inference_compute_output_duration_us counter
nv_inference_compute_output_duration_us{model="densenet_onnx",version="1"} 0
nv_inference_compute_output_duration_us{model="cc_fraud_detect_model",version="1"} 0
# HELP nv_inference_pending_request_count Instantaneous number of pending requests awaiting execution per-model.
# TYPE nv_inference_pending_request_count gauge
nv_inference_pending_request_count{model="densenet_onnx",version="1"} 0
nv_inference_pending_request_count{model="cc_fraud_detect_model",version="1"} 0
# HELP nv_pinned_memory_pool_total_bytes Pinned memory pool total memory size, in bytes
# TYPE nv_pinned_memory_pool_total_bytes gauge
nv_pinned_memory_pool_total_bytes 268435456
# HELP nv_pinned_memory_pool_used_bytes Pinned memory pool used memory size, in bytes
# TYPE nv_pinned_memory_pool_used_bytes gauge
nv_pinned_memory_pool_used_bytes 0
# HELP nv_cpu_utilization CPU utilization rate [0.0 - 1.0]
# TYPE nv_cpu_utilization gauge
nv_cpu_utilization 0.02255639097744361
# HELP nv_cpu_memory_total_bytes CPU total memory (RAM), in bytes
# TYPE nv_cpu_memory_total_bytes gauge
nv_cpu_memory_total_bytes 16861294592
# HELP nv_cpu_memory_used_bytes CPU used memory (RAM), in bytes
# TYPE nv_cpu_memory_used_bytes gauge
nv_cpu_memory_used_bytes 4000546816
```
</details>

Sample gRPC Health Check
<details>
  <summary>Click to expand the shell command</summary>

<br>This is a simple gRPC health Check achieved through port-forward of openshift pod <br>
For this code snippet to work, please remove grpcConfig.tlsSpec.tlsSecretName from the Crd configuration <br>
<b>Note:This configuration are for liveness checkness only.Please don't use it in production</b>

```sh
[root@a3elp45 OpenShift-AI-Toolkit-Operator]# oc get pod | grep triton
triton-server-fa7c1cee-triton-pvc-5f75fb5679-kx4nn         1/1     Running   0             5s
```

Copy the pod name and use it in the next command<br>
```sh
oc port-forward pod/<pod-name-from-command-above> 8001:8001 #Execute in a new shell
```

Run the health check script
```sh
(base) shivanggoswami@Shivangs-MacBook-Pro operator-examples % python grpc-server-liveness.py
Triton server is live.
```
</details>
