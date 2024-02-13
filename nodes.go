package main

import (
	"context"
	"fmt"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func updateNodesData(metricsClient *metricsv.Clientset, clientset *kubernetes.Clientset, nodeTable *widgets.Table) {
	nodeTable.Rows = [][]string{
		{"Node", "CPU Usage (Cores)", "Memory Usage (MiB)"},
	}

	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(context.Background(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Fetch node information
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	nodeInfo := make(map[string]corev1.Node)
	for _, node := range nodes.Items {
		nodeInfo[node.Name] = node
	}

	for _, metrics := range nodeMetricsList.Items {
		cpuUsage := metrics.Usage[corev1.ResourceCPU]
		cpuUsageCores := float64(cpuUsage.MilliValue()) / 1000
		memoryUsage := metrics.Usage[corev1.ResourceMemory]
		memoryUsageGib := float64(memoryUsage.Value()) / 1024 / 1024 / 1024

		// Fetch the total CPU and memory for the node
		totalCPU := nodeInfo[metrics.Name].Status.Allocatable[corev1.ResourceCPU]
		totalMemory := nodeInfo[metrics.Name].Status.Allocatable[corev1.ResourceMemory]
		totalCPUCores := float64(totalCPU.MilliValue()) / 1000
		totalMemoryGib := float64(totalMemory.Value()) / 1024 / 1024 / 1024

		// Calculate the percentage of CPU and memory used
		cpuPercentage := (cpuUsageCores / totalCPUCores) * 100
		memoryPercentage := (memoryUsageGib / totalMemoryGib) * 100

		// Convert percentages to a text-based bar representation
		cpuBar := renderBar(cpuPercentage, 20) // 20 is the bar width in characters
		memoryBar := renderBar(memoryPercentage, 20)

		cpuCell := fmt.Sprintf("%s %.3f/%.0f", cpuBar, cpuUsageCores, totalCPUCores)
		memoryCell := fmt.Sprintf("%s %.3f/%.3f", memoryBar, memoryUsageGib, totalMemoryGib)

		nodeTable.Rows = append(nodeTable.Rows, []string{
			metrics.Name,
			cpuCell,
			memoryCell,
		})

	}

	ui.Render(nodeTable)

}
