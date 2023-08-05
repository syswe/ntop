package mainpackage

import (
	"flag"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

// package main

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	bc := widgets.NewBarChart()
	bc.Title = "Node Resource Usage"
	bc.SetRect(0, 0, 50, 10)

	// set labels and data
	labels := make([]string, len(nodeList.Items))
	data := make([]float64, len(nodeList.Items))
	for i, node := range nodeList.Items {
		memory := node.Status.Capacity[corev1.ResourceMemory]
		memoryBytes := memory.Value()
		labels[i] = node.Name
		data[i] = float64(memoryBytes)
	}
	bc.Labels = labels
	bc.Data = data

	// colors
	bc.BarColors = []ui.Color{ui.ColorRed, ui.ColorYellow, ui.ColorGreen}
	bc.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorBlue)}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorYellow)}

	ui.Render(bc)

	// handle resize
	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				bc.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(bc)
			}
		}
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
