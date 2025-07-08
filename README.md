# Messaging app

## Requirements
* Be deployed via Kubernetes (deploy it locally on k3s and on test machine)
* Use local service discovery to detect and message all active pods
* Demonstrate reliable in-cluster communication

## Running locally
1. Install k3d
2. Create a cluster with 3 nodes (1 control pane, 2 workers)
3. Build images and deploy them to the cluster with deploy.sh
