package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	_ "strings"

	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	"github.com/olekukonko/tablewriter"
)

type PodInfo struct {
	PodName    string
	Containers []ContainerInfo
}

type ContainerInfo struct {
	Name string
	CPU  resource.Quantity
	Mem  resource.Quantity
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Cannot get home directory: %v", err)
	}

	kubeconfig := flag.String("kubeconfig", filepath.Join(homeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	podsOrderByCPU := flag.Bool("pods-order-by-cpu", false, "(optional) order pods by CPU usage instead of memory usage")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	metricsClientset, err := metricsv.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating metrics clientset: %v", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting nodes: %v", err)
	}

	for _, node := range nodes.Items {
		fmt.Println("Node: ", node.Name)
		metrics, err := metricsClientset.NodeMetricses().Get(context.TODO(), node.Name, v1.GetOptions{})
		if err != nil {
			re
			log.Fatalf("Error getting node metrics: %v", err)
		}
		// fmt.Printf("CPU Usage: %.3f cores\n", float64(metrics.Usage[corev1.ResourceCPU].MilliValue())/1000)
		// fmt.Printf("Memory Usage: %.3f Mi\n", float64(metrics.Usage[corev1.ResourceMemory].Value())/1024/1024)
		cpuUsage := metrics.Usage[corev1.ResourceCPU]
		cpuUsageMilli := cpuUsage.AsDec().UnscaledBig().Int64()
		fmt.Printf("CPU Usage: %.3f cores\n", float64(cpuUsageMilli)/1000)

		memoryUsage := metrics.Usage[corev1.ResourceMemory]
		memoryUsageMi := memoryUsage.ScaledValue(resource.Mega)
		fmt.Printf("Memory Usage: %.3f Mi\n", float64(memoryUsageMi))

	}

	pods, err := clientset.CoreV1().Pods("").List(context.Background(), v1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting pods: %v", err)
	}

	podInfos := make([]PodInfo, 0)
	for _, pod := range pods.Items {
		podMetrics, err := metricsClientset.PodMetricses(pod.Namespace).Get(context.TODO(), pod.Name, v1.GetOptions{})
		if err != nil {
			log.Printf("Error getting pod metrics: %v", err)
			continue
		}
		containerInfos := make([]ContainerInfo, len(podMetrics.Containers))
		for i, container := range podMetrics.Containers {
			containerInfos[i] = ContainerInfo{
				Name: container.Name,
				CPU:  container.Usage[corev1.ResourceCPU],
				Mem:  container.Usage[corev1.ResourceMemory],
			}
		}
		podInfos = append(podInfos, PodInfo{PodName: pod.Name, Containers: containerInfos})
	}

	sort.Slice(podInfos, func(i, j int) bool {
		if *podsOrderByCPU {
			return float64(podInfos[i].Containers[0].CPU.MilliValue()) > float64(podInfos[j].Containers[0].CPU.MilliValue())
		}
		return float64(podInfos[i].Containers[0].Mem.Value()) > float64(podInfos[j].Containers[0].Mem.Value())
	})

	// Only display top 10 pods
	data := make([][]string, 0)
	for i, podInfo := range podInfos {
		if i >= 10 {
			break
		}
		for _, container := range podInfo.Containers {
			data = append(data, []string{
				podInfo.PodName,
				container.Name,
				fmt.Sprintf("%.3f", float64(container.CPU.MilliValue())/1000),
				fmt.Sprintf("%.3f Mi", float64(container.Mem.Value())/1024/1024),
			})
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Pod", "Container", "CPU Usage (cores)", "Memory Usage (Mi)"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}
