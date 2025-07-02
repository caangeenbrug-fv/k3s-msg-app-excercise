#!/usr/bin/bash

# Build image
docker build . -t caangeenbrug-fv/msg-app:$1

# Make image available to K3s as it relies on containerd instead of Docker
docker save caangeenbrug-fv/msg-app:$1 | sudo k3s ctr images import -
