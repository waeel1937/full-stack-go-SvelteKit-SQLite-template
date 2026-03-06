package connector

import (
	"context"
	"log"
	"time"

	"edge-app/internal/core"
)

type OPCUANode struct {
	NodeID string `yaml:"node_id"`
	Key    string `yaml:"key"`
	Source string `yaml:"source"`
}

type OPCUAConfig struct {
	Enabled  bool       `yaml:"enabled"`
	Endpoint string     `yaml:"endpoint"`
	PollMs   int64      `yaml:"poll_ms"`
	Nodes    []OPCUANode `yaml:"nodes"`
}

type OPCUAConnector struct {
	Config OPCUAConfig
	Bus    *core.Bus
}

func (o *OPCUAConnector) Run() {
	if !o.Config.Enabled {
		log.Println("opcua connector disabled")
		return
	}
	log.Printf("opcua connector starting endpoint=%s poll=%dms nodes=%d",
		o.Config.Endpoint, o.Config.PollMs, len(o.Config.Nodes))

	poll := time.Duration(o.Config.PollMs) * time.Millisecond
	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	// NOTE: This is a template stub. To use real OPC-UA, add:
	//   go get github.com/gopcua/opcua
	// Then replace the poll loop below with actual OPC-UA client calls.
	//
	// Example with gopcua:
	//   c, _ := opcua.NewClient(o.Config.Endpoint)
	//   c.Connect(ctx)
	//   resp, _ := c.Read(&ua.ReadRequest{...})

	ctx := context.Background()
	_ = ctx

	for t := range ticker.C {
		for _, node := range o.Config.Nodes {
			// Stub: replace with real OPC-UA read
			log.Printf("opcua: would read node=%s key=%s (stub)", node.NodeID, node.Key)
			_ = t
			_ = node
		}
	}
}
