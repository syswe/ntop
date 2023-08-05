
import "log"

package main

import (
"log"
"strconv"

ui "github.com/gizak/termui/v3"
"github.com/gizak/termui/v3/widgets"
)

type NodeUsage struct {
	Name  string
	Usage int64
}

func main() {
	// your node usages, replace with real data
	nodeUsages := []NodeUsage{
		{Name: "node1", Usage: 12345},
		{Name: "node2", Usage: 12340},
		{Name: "node3", Usage: 12335},
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	bc := widgets.NewBarChart()
	bc.Title = "Node Resource Usage"
	bc.SetRect(0, 0, 50, 10)

	// set labels and data
	labels := make([]string, len(nodeUsages))
	data := make([]float64, len(nodeUsages))
	for i, nu := range nodeUsages {
		labels[i] = nu.Name
		data[i] = float64(nu.Usage)
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


