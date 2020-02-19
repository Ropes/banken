package view

import (
	"context"
	"fmt"
	"log"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func Init(ctx context.Context, n int) (topN, reqCnts, alerts *widgets.List) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	maxX, maxY := ui.TerminalDimensions()

	// TopN URL list
	topN = widgets.NewList()
	topN.Title = fmt.Sprintf("Top  %d HTTP Requested Paths", n)
	topN.Rows = []string{
		"[0] github.com/gizak/termui/v3",
		"[1] [你好，世界](fg:blue)",
		"[2] [こんにちは世界](fg:red)",
		"[3] [color](fg:white,bg:green) output",
		"[4] output.go",
		"[5] random_out.go",
		"[6] dashboard.go",
		"[7] foo",
		"[8] bar",
		"[9] baz",
	}
	topQuarter := maxX / 3
	topSplit := topQuarter * 2
	topN.TextStyle = ui.NewStyle(ui.ColorYellow)
	topN.WrapText = false
	topN.SetRect(0, 0, topSplit, n+2)

	// Req Avgs
	reqCnts = widgets.NewList()
	reqCnts.Title = "HTTP Request Interval Counts"
	reqCnts.Rows = []string{}
	reqCnts.TextStyle = ui.NewStyle(ui.ColorBlue)
	reqCnts.SetRect(topSplit+1, 0, maxX, n+2)

	// Alert List
	alerts = widgets.NewList()
	alerts.Title = "HTTP Req Rate Alerts"
	alerts.Rows = []string{}
	alerts.TextStyle = ui.NewStyle(ui.ColorRed)
	alerts.WrapText = true
	alerts.SetRect(0, n+3, maxX, maxY-(n+3))

	ui.Render(topN)
	ui.Render(reqCnts)
	ui.Render(alerts)

	return topN, reqCnts, alerts
}

func Run(can context.CancelFunc, topN, reqCnts, alerts *widgets.List) {
	defer ui.Close()
	// Alert list scrolling hooks
	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		ui.Render(topN)
		ui.Render(reqCnts)
		ui.Render(alerts)

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
