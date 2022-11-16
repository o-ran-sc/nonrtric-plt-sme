#!/bin/sh

kubectl create -f configs/chartmuseum.yaml
kubectl wait deployment -n default chartmuseum-deployment --for=condition=available --timeout=90s
kubectl create -f configs/capif.yaml
