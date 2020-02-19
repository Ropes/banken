// Package view provides TermUI definitions for displaying collected http
// traffic statistics to the running terminal.
package view

import (
	"context"
	"fmt"
	"log"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// Init constructs termui UI data structures and returns them so data
// controllers can update the UI.
func Init(ctx context.Context, n int) (topN, reqCnts, alerts *widgets.List) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	maxX, maxY := ui.TerminalDimensions()
	title := widgets.NewParagraph()
	minY := 3
	title.SetRect(0, 0, maxX, minY)
	title.Text = "Banken[番犬] HTTP Traffic Monitor -- press 'q' to quit"
	title.TextStyle = ui.NewStyle(ui.ColorCyan)
	ui.Render(title)

	// TopN URL list
	topN = widgets.NewList()
	topN.Title = fmt.Sprintf("Top %d HTTP Requested Paths", n)
	topN.Rows = []string{}
	topQuarter := maxX / 3
	topSplit := topQuarter * 2
	topN.TitleStyle = ui.NewStyle(ui.ColorYellow)
	topN.WrapText = false
	topN.SetRect(0, minY, topSplit, minY+n+2)

	// Req Avgs
	reqCnts = widgets.NewList()
	reqCnts.Title = "HTTP Requests per Timepspan"
	reqCnts.Rows = []string{}
	reqCnts.TitleStyle = ui.NewStyle(ui.ColorBlue)
	reqCnts.WrapText = false
	reqCnts.SetRect(topSplit+1, minY, maxX, minY+n+2)

	// Alert List
	alerts = widgets.NewList()
	alerts.Title = "HTTP Req Rate Alerts"
	alerts.Rows = []string{}
	alerts.SelectedRowStyle = ui.NewStyle(ui.ColorRed)
	alerts.TitleStyle = ui.NewStyle(ui.ColorRed)
	alerts.WrapText = true
	alerts.SetRect(0, minY+n+3, maxX, maxY-(n+3))

	ui.Render(topN)
	ui.Render(reqCnts)
	ui.Render(alerts)

	return topN, reqCnts, alerts
}

// Run catches key events which are needed for scrolling Alert notices in the
// UI, and catching shutdown commands. Calling can context.CancelFunc() signals
// the controllers to exit by closing the main context.Context.
func Run(can context.CancelFunc, topN, reqCnts, alerts *widgets.List) {
	defer ui.Close()
	// Alert list scrolling hooks
	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		alertsScrollable := false
		if len(alerts.Rows) > 0 {
			alertsScrollable = true
		}

		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			can()
			return
		case "j", "<Down>":
			if alertsScrollable {
				alerts.ScrollDown()
			}
		case "k", "<Up>":
			if alertsScrollable {
				alerts.ScrollUp()
			}
		case "<C-d>":
			if alertsScrollable {
				alerts.ScrollHalfPageDown()
			}
		case "<C-u>":
			if alertsScrollable {
				alerts.ScrollHalfPageUp()
			}
		case "<C-f>":
			if alertsScrollable {
				alerts.ScrollPageDown()
			}
		case "<C-b>":
			if alertsScrollable {
				alerts.ScrollPageUp()
			}
		case "g":
			if previousKey == "g" && alertsScrollable {
				alerts.ScrollTop()
			}
		case "<Home>":
			if alertsScrollable {
				alerts.ScrollTop()
			}
		case "G", "<End>":
			if alertsScrollable {
				alerts.ScrollBottom()
			}
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(topN)
		ui.Render(reqCnts)
		ui.Render(alerts)
	}
}
