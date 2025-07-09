# Messaging App Deployment Guide

This guide walks you through deploying your messaging app on Kubernetes, both locally (using k3d) and on a test machine. It highlights service discovery, reliable communication, and easy local development.

## How it works
Each pod sends some messages to its own service when it starts. When it retrieves a message, it keeps track of its whereabouts and propagates it through to the next pod. This pod is discovered by retrieving all pod IPs from a headless service. Then these IPs are sorted from low to high. The next pod is selected by taking the next IP in the sorted list from the current IP. After this process, the message is sent directly to the next pod using that IP.

You can verify its workings by watching the pod logs. They should look something like this:
```
received JSON from pod 'msg-app-deployment-6wvq5': {Message:Sending message from pod 10.42.2.44
 Trace:[msg-app-deployment-hw5pt msg-app-deployment-6wvq5 msg-app-deployment-6rd94 msg-app-deployment-hw5pt msg-app-deployment-6wvq5 msg-app-deployment-6rd94 msg-app-deployment-hw5pt msg-app-deployment-6wvq5 msg-app-deployment-6rd94 msg-app-deployment-hw5pt] SenderPodIP:10.42.2.44}
```
Note that the trace contains the same order of pods due to the IP sorting explained earlier.

## Requirements

- **Kubernetes Deployment**: Deploy the app on Kubernetes (locally with k3d/k3s and on a test machine).
- **Local Service Discovery**: The app should automatically discover and communicate with all active pods.
- **Reliable In-Cluster Messaging**: Ensure robust communication between pods within the cluster.

## Local Setup Instructions

### 1. Prerequisites

- **Install k3d**
  k3d is a lightweight wrapper to run k3s (Rancher's minimal Kubernetes) in Docker.
  Install it by following the official instructions for your OS.

- **Docker**
  Ensure Docker is installed and running on your machine.

- **Helm**
  Also ensure that Helm is installed and running on your machine.

### 2. Create a Kubernetes Cluster

Set up a cluster named `messaging-app-exercise` with 1 control plane and 3 worker nodes, exposing HTTP and HTTPS ports for ingress:

```sh
k3d cluster create messaging-app-exercise \
  --servers 1 \
  --agents 3 \
  --port 80:80@loadbalancer \
  --port 443:443@loadbalancer \
  --port 8000:8000@loadbalancer \
  --k3s-arg "--disable=traefik@server:0"
```

### 3. Install Kubernetes dependencies

Use Helm to install Kubernetes dependencies like so.
```sh
helm install traefik traefik/traefik -f values.yaml --wait
```

### 4. Build and Deploy the Application

- **Build Docker Images**
  Build your application images locally.

- **Deploy to the Cluster**
  Use the provided deployment script to build and deploy:

  ```sh
  ./local-deploy.sh
  ```

  This script handles image building and deployment to your local k3d cluster.

### 5. Accessing the App

- Map any required hostnames (such as those used in Ingress) to `127.0.0.1` in your `/etc/hosts` file if needed.
- Access the app via your browser or with tools like `curl` at `http://localhost` or the configured hostname.

## Key Features Demonstrated

- **Automatic Pod Discovery**: The app uses Kubernetes service discovery to find and message all active pods.
- **Reliable Messaging**: In-cluster communication is robust and fault-tolerant, ensuring messages are delivered between pods.

## Next Steps

- For test machine deployment, repeat the above steps on your target machine with k3s or any compatible Kubernetes environment.
- Adjust deployment scripts and manifests as needed for your specific environment or CI/CD pipeline.
