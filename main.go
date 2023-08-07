package main

import (
	"context"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/olekukonko/tablewriter"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Cannot get home directory: %v", err)
	}
	// kubeconfig := flag.String("kubeconfig", "/Users/abcd/.kube/config", "location to your kubeconfig file")
	kubeconfig := flag.String("kubeconfig", filepath.Join(homeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	podsOrder := flag.String("pods-order-by", "memory", "order pods listing by 'cpu' or 'memory'")

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(context.Background(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Node", "CPU Usage (Cores)", "Memory Usage (GiB)"})

	for _, metrics := range nodeMetricsList.Items {
		cpuUsage := metrics.Usage[corev1.ResourceCPU]
		cpuUsageCores := float64(cpuUsage.MilliValue()) / 1000

		memoryUsage := metrics.Usage[corev1.ResourceMemory]
		memoryUsageGib := float64(memoryUsage.Value()) / 1024 / 1024 / 1024

		table.Append([]string{metrics.Name, fmt.Sprintf("%.3f", cpuUsageCores), fmt.Sprintf("%.3f", memoryUsageGib)})
	}

	table.Render()

	pods, err := clientset.CoreV1().Pods("").List(context.Background(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	type PodInfo struct {
		PodName       string
		ContainerName string
		CPUUsage      float64
		MemoryUsage   float64
	}

	var podInfoList []PodInfo

	for _, pod := range pods.Items {
		podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(pod.Namespace).Get(context.Background(), pod.Name, v1.GetOptions{})
		if err != nil {
			fmt.Println("Error getting pod metrics:", err)
			continue
		}

		for _, container := range podMetrics.Containers {
			cpuUsage := container.Usage[corev1.ResourceCPU]
			cpuUsageCores := float64(cpuUsage.MilliValue()) / 1000

			memoryUsage := container.Usage[corev1.ResourceMemory]
			memoryUsageMib := float64(memoryUsage.Value()) / 1024 / 1024

			podInfoList = append(podInfoList, PodInfo{pod.Name, container.Name, cpuUsageCores, memoryUsageMib})
		}
	}

	switch *podsOrder {
	case "cpu":
		sort.Slice(podInfoList, func(i, j int) bool { return podInfoList[i].CPUUsage > podInfoList[j].CPUUsage })
	case "memory":
		sort.Slice(podInfoList, func(i, j int) bool { return podInfoList[i].MemoryUsage > podInfoList[j].MemoryUsage })
	}

	if len(podInfoList) > 10 {
		podInfoList = podInfoList[:10]
	}

	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Pod", "Container", "CPU Usage (Cores)", "Memory Usage (MiB)"})

	for _, podInfo := range podInfoList {
		table.Append([]string{podInfo.PodName, podInfo.ContainerName, fmt.Sprintf("%.3f", podInfo.CPUUsage), fmt.Sprintf("%.3f", podInfo.MemoryUsage)})
	}

	table.Render()
}
