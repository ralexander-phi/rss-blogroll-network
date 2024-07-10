package main

import (
	"encoding/json"
	"os"
	"slices"
)

type VisualData struct {
	Nodes []VizNode `json:"nodes"`
	Links []VizLink `json:"links"`
}

type VizNode struct {
	ID    string `json:"id"`
	Group int    `json:"group"`
}

type VizLink struct {
	Source      string `json:"source"`
	Destination string `json:"target"`
}

func (a *Analysis) Visualize() {
	viz := VisualData{}

	nodes := map[string]int{}

	rows := []Link{}
	result := a.db.Find(&rows)
	ohno(result.Error)

	for _, row := range rows {
		viz.Links = append(viz.Links, VizLink{
			Source:      row.SourceUrl,
			Destination: row.DestinationUrl,
		})
		nodes[row.SourceUrl] = row.SourceType
		nodes[row.DestinationUrl] = row.DestinationType
	}

	for node, nodeType := range nodes {
		viz.Nodes = append(viz.Nodes, VizNode{
			ID:    node,
			Group: nodeType,
		})
	}

	slices.SortFunc(viz.Nodes, func(a, b VizNode) int {
		if a.ID < b.ID {
			return -1
		} else if a.ID > b.ID {
			return 1
		} else {
			return 0
		}
	})

	slices.SortFunc(viz.Links, func(a, b VizLink) int {
		if a.Source < b.Source {
			return -1
		} else if a.Source > b.Source {
			return 1
		} else {
			if a.Destination < b.Destination {
				return -1
			} else if a.Destination > b.Destination {
				return 1
			} else {
				return 0
			}
		}
	})

	output, err := json.MarshalIndent(viz, "", "    ")
	ohno(err)
	err = os.WriteFile("static/index.json", output, os.FileMode(int(0660)))
	ohno(err)
}
