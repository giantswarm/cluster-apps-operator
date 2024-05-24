#!/bin/bash

apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)" --install-operators=false

KUBE_CONFIG=$(kind get kubeconfig) kubectl apply -f "https://raw.githubusercontent.com/kubernetes-sigs/cluster-api/v1.5.0/config/crd/bases/cluster.x-k8s.io_clusters.yaml"
