package connector

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"edge-app/internal/core"
)

type ModbusRegister struct {
	Address  uint16 `yaml:"address"`
	Key      string `yaml:"key"`
	Source   string `yaml:"source"`
	Type     string `yaml:"type"`
	Scale    float64 `yaml:"scale"`
}

type ModbusConfig struct {
	Enabled   bool             `yaml:"enabled"`
	Host      string           `yaml:"host"`
	Port      int              `yaml:"port"`
	UnitID    byte             `yaml:"unit_id"`
	PollMs    int64            `yaml:"poll_ms"`
	Registers []ModbusRegister `yaml:"registers"`
}

type ModbusConnector struct {
	Config ModbusConfig
	Bus    *core.Bus
	conn   net.Conn
	txID   uint16
}

func (m *ModbusConnector) Run() {
	if !m.Config.Enabled {
		log.Println("modbus connector disabled")
		return
	}
	log.Printf("modbus connector starting %s:%d unit=%d poll=%dms",
		m.Config.Host, m.Config.Port, m.Config.UnitID, m.Config.PollMs)

	poll := time.Duration(m.Config.PollMs) * time.Millisecond
	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	for t := range ticker.C {
		if m.conn == nil {
			addr := fmt.Sprintf("%s:%d", m.Config.Host, m.Config.Port)
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				log.Printf("modbus connect error: %v", err)
				continue
			}
			m.conn = conn
			log.Printf("modbus connected to %s", addr)
		}

		for _, reg := range m.Config.Registers {
			val, err := m.readRegister(reg)
			if err != nil {
				log.Printf("modbus read error reg=%d: %v", reg.Address, err)
				m.conn.Close()
				m.conn = nil
				break
			}
			m.Bus.PublishMetric(core.MetricEvent{
				Time:   t,
				Source: reg.Source,
				Key:    reg.Key,
				Value:  val,
				OK:     true,
			})
		}
	}
}

func (m *ModbusConnector) readRegister(reg ModbusRegister) (float64, error) {
	m.txID++
	count := uint16(1)
	if reg.Type == "float32" {
		count = 2
	}

	req := make([]byte, 12)
	binary.BigEndian.PutUint16(req[0:2], m.txID)
	binary.BigEndian.PutUint16(req[2:4], 0)
	binary.BigEndian.PutUint16(req[4:6], 6)
	req[6] = m.Config.UnitID
	req[7] = 3
	binary.BigEndian.PutUint16(req[8:10], reg.Address)
	binary.BigEndian.PutUint16(req[10:12], count)

	m.conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := m.conn.Write(req); err != nil {
		return 0, err
	}

	header := make([]byte, 9)
	if _, err := m.conn.Read(header); err != nil {
		return 0, err
	}

	byteCount := int(header[8])
	data := make([]byte, byteCount)
	if _, err := m.conn.Read(data); err != nil {
		return 0, err
	}

	scale := reg.Scale
	if scale == 0 {
		scale = 1.0
	}

	switch reg.Type {
	case "float32":
		if len(data) < 4 {
			return 0, fmt.Errorf("not enough data for float32")
		}
		bits := binary.BigEndian.Uint32(data[0:4])
		return float64(math.Float32frombits(bits)) * scale, nil
	case "int16":
		if len(data) < 2 {
			return 0, fmt.Errorf("not enough data for int16")
		}
		return float64(int16(binary.BigEndian.Uint16(data[0:2]))) * scale, nil
	default:
		if len(data) < 2 {
			return 0, fmt.Errorf("not enough data for uint16")
		}
		return float64(binary.BigEndian.Uint16(data[0:2])) * scale, nil
	}
}
