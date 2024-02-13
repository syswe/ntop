package main

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

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
