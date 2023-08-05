package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ResourceUsage struct {
	Name  string
	Usage int64
}

func main() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	clientset, _ := kubernetes.NewForConfig(config)

	// Create a context
	ctx := context.Background()

	// List Nodes
	nodes, _ := clientset.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	var nodeUsages []ResourceUsage
	for _, node := range nodes.Items {
		// Use node to get the resource usages.
		usage := getResourceUsage(node) // Replace this with your actual implementation
		nodeUsages = append(nodeUsages, ResourceUsage{Name: node.Name, Usage: usage})
	}

	// Sort nodes by usage
	sort.Slice(nodeUsages, func(i, j int) bool {
		return nodeUsages[i].Usage > nodeUsages[j].Usage
	})

	// Print nodes table
	fmt.Println("Nodes:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tUsage")
	for _, nu := range nodeUsages {
		fmt.Fprintf(w, "%s\t%d\n", nu.Name, nu.Usage)
	}
	w.Flush()
	fmt.Println()

	// List Pods
	pods, _ := clientset.CoreV1().Pods("").List(ctx, v1.ListOptions{})
	var podUsages []ResourceUsage
	for _, pod := range pods.Items {
		// Use pod to get the resource usages and store in ResourceUsage slice.
		usage := getPodUsage(pod) // Replace this with your actual implementation
		podUsages = append(podUsages, ResourceUsage{Name: pod.Name, Usage: usage})
	}

	// Sort pods by usage and get top 10
	sort.Slice(podUsages, func(i, j int) bool {
		return podUsages[i].Usage > podUsages[j].Usage
	})
	if len(podUsages) > 10 {
		podUsages = podUsages[:10]
	}

	// Print pods table
	fmt.Println("Top 10 Pods:")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tUsage")
	for _, pu := range podUsages {
		fmt.Fprintf(w, "%s\t%d\n", pu.Name, pu.Usage)
	}
	w.Flush()
}

// Placeholder functions, replace with your actual implementation

func getResourceUsage(node corev1.Node) int64 {
	// Get resource usage of the node
	return 0
}

func getPodUsage(pod corev1.Pod) int64 {
	// Get resource usage of the pod
	return 0
}
