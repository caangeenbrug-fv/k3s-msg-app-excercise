#!/usr/bin/bash

# Build image
docker build . -t caangeenbrug-fv/msg-app:latest

# Make image available to k3s
docker save caangeenbrug-fv/msg-app:latest | sudo k3s ctr images import -
