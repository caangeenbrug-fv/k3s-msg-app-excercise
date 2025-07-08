#!/usr/bin/bash

# Build image
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag flikweertvision/msg-app-excercise:latest \
    .

k3d kubeconfig merge exercise-cluster --kubeconfig-switch-context

# # Make image available to K3s as it relies on containerd instead of Docker
k3d image import flikweertvision/msg-app-excercise:latest -c exercise-cluster

# kubectl replace --force -f deployment.yaml
