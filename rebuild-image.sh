#!/usr/bin/bash

# Build image
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag flikweertvision/msg-app-excercise:$1 \
    .

# Make image available to K3s as it relies on containerd instead of Docker
docker save flikweertvision/msg-app-excercise:$1 | sudo k3s ctr images import -
