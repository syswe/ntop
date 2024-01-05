package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Cannot get home directory: %v", err)
	}
	// kubeconfig := flag.String("kubeconfig", "/Users/abcd/.kube/config", "location to your kubeconfig file")
	kubeconfig := flag.String("kubeconfig", filepath.Join(homeDir, ".kube", "config-10"), "(optional) absolute path to the kubeconfig file")
	// podsOrder := flag.String("pods-order-by", "memory", "order pods listing by 'cpu' or 'memory'")
	// var verbose = flag.Bool("verbose", false, "enable verbose output")

	flag.Parse() // parse the flags

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err.Error())
	// }

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

	// table := tablewriter.NewWriter(os.Stdout)
	// table.SetHeader([]string{"Node", "CPU Usage (Cores)", "Memory Usage (GiB)"})

	// for _, metrics := range nodeMetricsList.Items {
	// 	cpuUsage := metrics.Usage[corev1.ResourceCPU]
	// 	cpuUsageCores := float64(cpuUsage.MilliValue()) / 1000

	// 	memoryUsage := metrics.Usage[corev1.ResourceMemory]
	// 	memoryUsageGib := float64(memoryUsage.Value()) / 1024 / 1024 / 1024

	// 	table.Append([]string{metrics.Name, fmt.Sprintf("%.3f", cpuUsageCores), fmt.Sprintf("%.3f", memoryUsageGib)})
	// }

	// table.Render()

	// pods, err := clientset.CoreV1().Pods("").List(context.Background(), v1.ListOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }

	// type PodInfo struct {
	// 	PodName       string
	// 	ContainerName string
	// 	Namespace     string
	// 	CPUUsage      float64
	// 	MemoryUsage   float64
	// }

	// var podInfoList []PodInfo
	// fmt.Println("We're calculating your all pods in your cluster. Please wait.")
	// done := make(chan bool)
	// go showLoadingBar(done)

	// for _, pod := range pods.Items {
	// 	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(pod.Namespace).Get(context.Background(), pod.Name, v1.GetOptions{})
	// 	if err != nil {
	// 		if *verbose {
	// 			fmt.Printf("Error getting pod metrics for %s: %v\n", pod.Name, err)
	// 		}
	// 		continue
	// 	}

	// 	for _, container := range podMetrics.Containers {
	// 		cpuUsage := container.Usage[corev1.ResourceCPU]
	// 		cpuUsageCores := float64(cpuUsage.MilliValue()) / 1000

	// 		memoryUsage := container.Usage[corev1.ResourceMemory]
	// 		memoryUsageMib := float64(memoryUsage.Value()) / 1024 / 1024

	// 		podInfoList = append(podInfoList, PodInfo{pod.Name, container.Name, pod.Namespace, cpuUsageCores, memoryUsageMib})
	// 	}
	// }

	// done <- true // Signal to stop the loading bar
	// fmt.Println("\nCalculations Complete")

	// switch *podsOrder {
	// case "cpu":
	// 	sort.Slice(podInfoList, func(i, j int) bool { return podInfoList[i].CPUUsage > podInfoList[j].CPUUsage })
	// case "memory":
	// 	sort.Slice(podInfoList, func(i, j int) bool { return podInfoList[i].MemoryUsage > podInfoList[j].MemoryUsage })
	// }

	// if len(podInfoList) > 10 {
	// 	podInfoList = podInfoList[:10]
	// }

	// table = tablewriter.NewWriter(os.Stdout)
	// table.SetHeader([]string{"Pod", "Container", "Namespace", "CPU Usage (Cores)", "Memory Usage (MiB)"})

	// for _, podInfo := range podInfoList {
	// 	table.Append([]string{podInfo.PodName, podInfo.ContainerName, podInfo.Namespace, fmt.Sprintf("%.3f", podInfo.CPUUsage), fmt.Sprintf("%.3f", podInfo.MemoryUsage)})
	// }

	// table.Render()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Create a slice to hold the bar charts
	// var barCharts []*widgets.BarChart

	// Create table for node metrics
	nodeTable := widgets.NewTable()
	nodeTable.TextStyle = ui.NewStyle(ui.ColorWhite)
	nodeTable.RowSeparator = true
	nodeTable.BorderStyle = ui.NewStyle(ui.ColorGreen)
	//nodeTable.SetRect(0, 0, 100, 30)
	baseHeight := 0      // Minimum height of the table
	heightIncrement := 1 // Additional height per node
	maxHeight := 25      // Maximum height of the table

	nodeCount := nodeMetricsList.Size() // Get the count of nodes

	dynamicHeight := 0 // Dynamic height of the table
	if nodeCount > 5 {
		dynamicHeight = (baseHeight + (nodeCount * heightIncrement)) - 5 // Calculate dynamic height
	} else {
		dynamicHeight = baseHeight + (nodeCount * heightIncrement) // Calculate dynamic height
	}

	// Ensure the dynamic height does not exceed the maximum height
	if dynamicHeight > maxHeight {
		dynamicHeight = maxHeight
	}

	// Set the table dimensions
	nodeTable.SetRect(0, 0, 130, dynamicHeight)

	updateData := func() {
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

	updateData()

	// Ticker for updating data
	ticker := time.NewTicker(10 * time.Millisecond)
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
			updateData()
		}
	}

}

func renderBar(percentage float64, width int) string {
	filled := int((percentage / 100) * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

func showLoadingBar(done chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			fmt.Print(".")
			time.Sleep(500 * time.Millisecond) // Adjust the speed as needed
		}
	}
}
