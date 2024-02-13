package main

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
	parseFlags()

	const baseHeight = 0      // Minimum height of the table
	const heightIncrement = 1 // Additional height per node or pod
	const maxHeight = 30      // Maximum height of the table

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Failed to build kubeconfig: %v", err)
	}

	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(context.Background(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Initialize Kubernetes clientset for node information
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Initialize termui
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Create a slice to hold the bar charts
	// var barCharts []*widgets.BarChart

	nodeTable := widgets.NewTable()
	// Configure nodeTable...

	podsTable := widgets.NewTable()
	// Default configuration for podsTable...

	if *podsFlag {
		// Additional configuration specific to podsTable
		podsTable.TextStyle = ui.NewStyle(ui.ColorWhite)
		podsTable.RowSeparator = true
		podsTable.BorderStyle = ui.NewStyle(ui.ColorGreen)
		podsTable.Title = "Pods"
		dynamicHeight := ((min(*countPods, maxHeight) + baseHeight) * 3) - 6 // Adjust baseHeight and maxHeight as needed
		podsTable.SetRect(0, 0, 160, dynamicHeight)

		updatePodsData(metricsClient, clientset, podsTable, *countPods)
	} else {
		nodeTable.TextStyle = ui.NewStyle(ui.ColorWhite)
		nodeTable.RowSeparator = true
		nodeTable.BorderStyle = ui.NewStyle(ui.ColorGreen)
		nodeTable.Title = "Nodes"
		nodeCount := len(nodeMetricsList.Items) // Get the count of nodes
		dynamicHeight := 0                      // Dynamic height of the table
		if nodeCount > 5 {
			dynamicHeight = ((baseHeight + (nodeCount * heightIncrement)) * 3) - 7 // Calculate dynamic height
		} else {
			dynamicHeight = baseHeight + (nodeCount * heightIncrement) // Calculate dynamic height
		}
		if dynamicHeight > maxHeight {
			dynamicHeight = maxHeight
		}
		fmt.Println("Dynamic Height:", dynamicHeight)
		nodeTable.SetRect(0, 0, 160, dynamicHeight)

		updateNodesData(metricsClient, clientset, nodeTable)
	}

	if *podsFlag {
		updatePodsData(
			metricsClient,
			clientset,
			podsTable,
			*countPods,
		)
	} else {
		updateNodesData(
			metricsClient,
			clientset,
			nodeTable,
		)
	}

	// Ticker for updating data
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	// Handle key q to quit
	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			if e.ID == "q" || e.Type == ui.KeyboardEvent {
				return
			}
		case <-ticker.C:
			if *podsFlag {
				updatePodsData(
					metricsClient,
					clientset,
					nodeTable,
					*countPods,
				)
			} else {
				updateNodesData(
					metricsClient,
					clientset,
					nodeTable,
				)
			}
		}
	}
}
