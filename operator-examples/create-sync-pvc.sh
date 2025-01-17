#!/bin/bash

# Start the timer
start_time=$(date +%s)

# Set your variables
LOCAL_DIR="./models"    # Local directory to copy data from
NAMESPACE="test"        # Kubernetes namespace
PVC_NAME="triton-pvc"   # PVC name
POD_NAME="alpine-pvc-pod" # Pod name
MOUNT_PATH="/mnt/data"  # PVC mount path inside the pod
CONTAINER_NAME="alpine" # Container name
PVC_STORAGE="10Gi"      # PVC storage size
ALPINE_IMAGE="alpinelinux/rsyncd:latest-s390x" # Alpine image to use for the pod
CLEAN=true             # Set to true to clean (delete) existing PVC and Pod before execution

# Check if the local directory exists
if [ ! -d "$LOCAL_DIR" ]; then
    echo "Local directory $LOCAL_DIR does not exist."
    exit 1
fi

# Check if the namespace exists and create it if it doesn't
if ! oc get namespace "$NAMESPACE" &>/dev/null; then
  echo "Namespace '$NAMESPACE' does not exist. Creating..."
  oc create namespace "$NAMESPACE"
else
  echo "Namespace '$NAMESPACE' already exists."
fi

# If clean is true, delete existing PVC and Pod
if [ "$CLEAN" = true ]; then
    echo "Cleaning existing PVC and Pod..."
    
    # Delete the pod if it exists
    if oc get pod "$POD_NAME" -n "$NAMESPACE" &>/dev/null; then
        echo "Deleting existing pod $POD_NAME..."
        oc delete pod "$POD_NAME" -n "$NAMESPACE"
    fi
    
    # Delete the PVC if it exists
    if oc get pvc "$PVC_NAME" -n "$NAMESPACE" &>/dev/null; then
        echo "Deleting existing PVC $PVC_NAME..."
        oc delete pvc "$PVC_NAME" -n "$NAMESPACE"
    fi
fi

# Check if the PVC already exists, if not, create it
PVC_EXISTS=$(oc get pvc "$PVC_NAME" -n "$NAMESPACE" --ignore-not-found)
if [ -z "$PVC_EXISTS" ]; then
    # Create PVC if it doesn't exist
    echo "Creating PVC $PVC_NAME..."
    oc apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: $PVC_NAME
  namespace: $NAMESPACE
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: $PVC_STORAGE
EOF
    # Check PVC status and wait for binding
    STATUS=$(oc get pvc "$PVC_NAME" -n "$NAMESPACE" -o jsonpath='{.status.phase}')
    if [ "$STATUS" == "Bound" ]; then
        echo "PVC is already bound, no need to wait."
    else
        # Wait for PVC to be bound
        echo "Waiting for PVC to be bound..."
        oc wait --for=condition=Bound pvc/$PVC_NAME -n "$NAMESPACE" --timeout=120s
    fi
else
    echo "PVC $PVC_NAME already exists."
fi

# Create the Pod with Alpine image and mount the PVC
echo "Creating pod $POD_NAME with Alpine image and PVC mount..."
oc apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: $POD_NAME
  namespace: $NAMESPACE
spec:
  containers:
  - name: $CONTAINER_NAME
    image: $ALPINE_IMAGE
    command: ["sleep", "3600"]  # Keeps the pod running
    volumeMounts:
    - mountPath: $MOUNT_PATH
      name: alpine-pvc
  volumes:
  - name: alpine-pvc
    persistentVolumeClaim:
      claimName: $PVC_NAME
EOF

# Wait for the pod to be running
echo "Waiting for pod $POD_NAME to start..."
# Check if the pod is already in the 'Ready' state
STATUS=$(oc get pod $POD_NAME -n $NAMESPACE -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')
if [ "$STATUS" == "True" ]; then
    echo "Pod $POD_NAME is already ready."
else
    # Wait for the pod to be ready
    echo "Waiting for pod $POD_NAME to be ready..."
    oc wait --for=condition=Ready pod/$POD_NAME -n "$NAMESPACE" --timeout=120s
fi


# If parallel flag is false, do a single rsync
echo "Copying data from $LOCAL_DIR to $MOUNT_PATH inside pod $POD_NAME..."
oc rsync --compress=true --progress=true -n $NAMESPACE "$LOCAL_DIR/" "$POD_NAME:$MOUNT_PATH"

# Verify if the rsync operation was successful
if [ $? -eq 0 ]; then
    echo "Data successfully copied to PVC."
else
    echo "Failed to copy data to PVC."
    exit 1
fi

# End the timer
end_time=$(date +%s)

# Calculate and display the execution time
execution_time=$((end_time - start_time))
echo "Script execution completed in $execution_time seconds."
