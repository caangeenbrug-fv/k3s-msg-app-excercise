#!/usr/bin/bash

# Build image
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag flikweertvision/msg-app-excercise:0.0.9 \
    .

# Make image available to K3s as it relies on containerd instead of Docker
docker save flikweertvision/msg-app-excercise:0.0.9 | sudo k3s ctr images import -

k3s kubectl apply -f deployment.yaml

k3s kubectl delete pods -l app=msg-app
