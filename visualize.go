package main

import (
	"encoding/json"
	"os"
)

type VisualData struct {
	Nodes []VizNode `json:"nodes"`
	Links []VizLink `json:"links"`
}

type VizNode struct {
	ID    string `json:"id"`
	Group int64  `json:"group"`
}

type VizLink struct {
	Source      string `json:"source"`
	Destination string `json:"target"`
}

func (a *Analysis) Visualize() {
	viz := VisualData{}

	nodes := map[string]int64{}
	linkRows, err := a.db.Query(`
    SELECT source_url, source_type, destination_url, destination_type
      FROM links;`,
	)
	ohno(err)
	for linkRows.Next() {
		var sourceUrl string
		var sourceType int64
		var destinationUrl string
		var destinationType int64
		err = linkRows.Scan(
			&sourceUrl,
			&sourceType,
			&destinationUrl,
			&destinationType,
		)
		ohno(err)

		viz.Links = append(viz.Links, VizLink{
			Source:      sourceUrl,
			Destination: destinationUrl,
		})
		nodes[sourceUrl] = sourceType
		nodes[destinationUrl] = destinationType
	}

	for node, nodeType := range nodes {
		viz.Nodes = append(viz.Nodes, VizNode{
			ID:    node,
			Group: nodeType,
		})
	}
	output, err := json.Marshal(viz)
	ohno(err)
	err = os.WriteFile("static/index.json", output, os.FileMode(int(0660)))
	ohno(err)
}
