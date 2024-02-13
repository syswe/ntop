package main

import (
	"flag"
	"os"
	"path/filepath"
)

var (
	kubeconfig *string
	podsFlag   *bool
	countPods  *int
)

func parseFlags() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Cannot get home directory: " + err.Error())
	}

	kubeconfig = flag.String("kubeconfig", filepath.Join(homeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	podsFlag = flag.Bool("pods", false, "if set, displays pods instead of nodes")
	countPods = flag.Int("countpods", 10, "number of pods to display")

	flag.Parse()
}
