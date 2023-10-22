package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

var configFromFile = flag.String("c", "", "if nonempty, load kea JSON config from file instead of querying unix domain socket")

const configQuery = `{"command": "config-get"}`

type ParsedKeaConfig struct {
	KeaConfig KeaConfig `json:"arguments"`
}

type KeaConfig struct {
	Dhcp4 Dhcp4 `json:"Dhcp4"`
}

type Dhcp4 struct {
	Subnets     []Subnet `json:"subnet4"`
	SubnetsByID map[uint64]string
}

type Subnet struct {
	ID      uint64 `json:"id"`
	Netname string `json:"subnet"`
}

func queryConfig() (*KeaConfig, error) {
	var c *KeaConfig
	var err error
	var rawJSON []byte
	if *jsonFromFile == "" {
		logger.Debug("Reading Kea config from socket", "path", *sockPath)
		rawJSON, err = queryKeaOnce(configQuery)
	} else {
		logger.Debug("Reading Kea config from file", "path", *configFromFile)
		rawJSON, err = getRawJSONFromFile(*configFromFile)
	}
	if err != nil {
		return nil, fmt.Errorf("could not query Kea for config: %w", err)
	}
	logger.Debug("Parsing JSON", "size", len(rawJSON))
	c, err = fromJSON(rawJSON)
	if err != nil {
		return nil, fmt.Errorf("could not parse Kea config: %w", err)
	}
	return c, err
}

func fromJSON(data []byte) (*KeaConfig, error) {
	var pkc ParsedKeaConfig
	err := json.Unmarshal(data, &pkc)
	if err != nil {
		return nil, err
	}
	c := pkc.KeaConfig
	c.Dhcp4.SubnetsByID = make(map[uint64]string)
	for _, sn := range c.Dhcp4.Subnets {
		c.Dhcp4.SubnetsByID[sn.ID] = sn.Netname
	}
	return &c, nil
}

func (c KeaConfig) subnetFromID(nettype int, id uint64) (string, error) {
	var subnet string
	var err error
	var ok bool

	switch nettype {
	case 4:
		subnet, ok = c.Dhcp4.SubnetsByID[id]
		if !ok {
			subnet = "unknown"
		}
	default:
		err = fmt.Errorf("unknown nettype '%d', want 4", nettype)
	}

	return subnet, err
}
