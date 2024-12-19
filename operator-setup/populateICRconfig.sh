#!/bin/bash

oc get secret/pull-secret -n openshift-config --template='{{index .data ".dockerconfigjson" | base64decode}}' > pull-secret.json

oc -n openshift-config create secret docker-registry icr-registry --docker-server=$REGISTRY_URL --docker-username=iamapikey --docker-password=$API_KEY --docker-email=$EMAIL

jq -s '.[0] * .[1]' pull-secret.json <(oc get secret -n openshift-config icr-registry -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d) > merged_config.json

oc delete secret -n openshift-config icr-registry 

oc set data secret/pull-secret -n openshift-config --from-file=.dockerconfigjson=merged_config.json

rm -f pull-secret.json merged_config.json