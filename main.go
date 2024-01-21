package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	kubeconfig := flag.String("kubeconfig", filepath.Join(homeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// podsOrder := flag.String("pods-order-by", "memory", "order pods listing by 'cpu' or 'memory'")
	// var verbose = flag.Bool("verbose", false, "enable verbose output")

	// list pods or nodes with flags --pods or --nodes, default list nodes if no flags
	var podsFlag = flag.Bool("pods", false, "if set, displays pods instead of nodes")
	var countPods = flag.Int("countpods", 10, "number of pods to display")

	flag.Parse() // parse the flags

	const baseHeight = 0      // Minimum height of the table
	const heightIncrement = 1 // Additional height per node or pod
	const maxHeight = 30      // Maximum height of the table

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

	fmt.Println("Node Metrics List Size:", len(nodeMetricsList.Items))
	fmt.Println("Dynamic Height:", (baseHeight + (len(nodeMetricsList.Items) * heightIncrement)))
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

func renderBar(percentage float64, width int) string {
	if percentage < 0 {
		percentage = 0
	} else if percentage > 100 {
		percentage = 100
	}

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

func calculatePercentage(current, limit float64) float64 {
	if limit == 0 {
		return 0 // Avoid division by zero; handle as needed
	}
	return (current / limit) * 100
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func updatePodsData(metricsClient *metricsv.Clientset, clientset *kubernetes.Clientset, podsTable *widgets.Table, countPods int) {
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), v1.ListOptions{})
	if err != nil {
		return
	}

	type PodMetricsInfo struct {
		Name           string
		Namespace      string
		CPUUsageCores  float64
		MemoryUsageMiB float64
	}

	var (
		podMetricsList []PodMetricsInfo
		mutex          sync.Mutex
	)

	for _, pod := range pods.Items {
		go func(pod corev1.Pod) {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // Random sleep to avoid API rate limiting
			podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(pod.Namespace).Get(context.Background(), pod.Name, v1.GetOptions{})
			if err != nil {
				return
			}

			var totalCPUUsage, totalMemoryUsage resource.Quantity
			for _, container := range podMetrics.Containers {
				totalCPUUsage.Add(container.Usage[corev1.ResourceCPU])
				totalMemoryUsage.Add(container.Usage[corev1.ResourceMemory])
			}
			cpuUsageCores := float64(totalCPUUsage.MilliValue()) / 1000
			memoryUsageMiB := float64(totalMemoryUsage.Value()) / 1024 / 1024

			mutex.Lock()
			podMetricsList = append(podMetricsList, PodMetricsInfo{
				Name:           pod.Name,
				Namespace:      pod.Namespace,
				CPUUsageCores:  cpuUsageCores,
				MemoryUsageMiB: memoryUsageMiB,
			})
			mutex.Unlock()
		}(pod)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			if e.ID == "q" || e.ID == "<C-c>" {
				return
			}
		case <-ticker.C:
			mutex.Lock()
			// Sort and limit the list to the specified number of pods
			sort.Slice(podMetricsList, func(i, j int) bool {
				return podMetricsList[i].MemoryUsageMiB > podMetricsList[j].MemoryUsageMiB
			})
			limitedPodsList := podMetricsList
			if len(podMetricsList) > countPods {
				limitedPodsList = podMetricsList[:countPods]
			}

			podsTable.Rows = [][]string{{"Pod", "Namespace", "CPU Usage (Cores)", "Memory Usage (MiB)"}}
			for _, podInfo := range limitedPodsList {
				cpuBar := renderBar(calculatePercentage(podInfo.CPUUsageCores, 100), 20)      // Assume 100 as a max for percentage calculation
				memoryBar := renderBar(calculatePercentage(podInfo.MemoryUsageMiB, 1000), 20) // Assume 1000 MiB as a max for percentage calculation

				podsTable.Rows = append(podsTable.Rows, []string{
					podInfo.Name,
					podInfo.Namespace,
					fmt.Sprintf("%.2f ", podInfo.CPUUsageCores) + cpuBar,
					fmt.Sprintf("%.2f ", podInfo.MemoryUsageMiB) + memoryBar,
				})
			}
			mutex.Unlock()
			ui.Render(podsTable)
		}
	}
}

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
