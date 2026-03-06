package connector

import (
	"log"

	"edge-app/internal/core"
)

type Config struct {
	Modbus ModbusConfig `yaml:"modbus"`
	OPCUA  OPCUAConfig  `yaml:"opcua"`
}

func StartAll(cfg Config, bus *core.Bus) {
	if cfg.Modbus.Enabled {
		m := &ModbusConnector{Config: cfg.Modbus, Bus: bus}
		go m.Run()
		log.Println("modbus connector launched")
	}
	if cfg.OPCUA.Enabled {
		o := &OPCUAConnector{Config: cfg.OPCUA, Bus: bus}
		go o.Run()
		log.Println("opcua connector launched")
	}
}
