package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const statsQuery = `{
  "command":"statistic-get-all",
  "arguments":{}
}`

var (
	sockPath     = flag.String("s", "/run/kea/kea4-ctrl-socket", "Path to Kea control socket")
	jsonFromFile = flag.String("f", "", "if nonempty, load stats JSON from file instead of querying unix domain socket")
)

func getStatsJSON() ([]byte, error) {
	var rawJSON []byte
	var err error
	if *jsonFromFile == "" {
		logger.Debug("Reading Kea stats from socket", "path", *sockPath)
		rawJSON, err = queryKeaOnce(statsQuery)
	} else {
		logger.Debug("Reading Kea stats from file", "path", *jsonFromFile)
		rawJSON, err = getRawJSONFromFile(*jsonFromFile)
	}
	if err != nil {
		return nil, fmt.Errorf("could not get raw JSON stats: %w", err)
	}
	return rawJSON, nil
}

func parseStats(rawJSON []byte) (*KeaCookedMetrics, error) {
	var stats KeaRawMetrics
	var err error
	err = json.Unmarshal(rawJSON, &stats)
	if err != nil {
		return nil, err
	}
	var cooked KeaCookedMetrics
	cooked.SubnetMetrics = make(map[uint64]KeaSubnetMetrics)
	for name, stat := range stats.Arguments {
		ml, ok := stat.([]interface{})
		if !ok {
			return nil, fmt.Errorf("stat is not an []interface{}: %#v", stat)
		}
		if strings.HasPrefix(name, "subnet[") {
			err = extractSubnetMetric(name, &cooked, ml)
			if err != nil {
				return nil, err
			}
		} else {
			val, err := getLatestMetricValue(ml)
			if err != nil {
				return nil, err
			}
			extractCookedMetrics(name, &cooked, val)
		}
	}
	return &cooked, nil
}

type KeaRawMetrics struct {
	Arguments map[string]interface{} `mapstructure:",remain"`
	Result    int                    `json:"result"`
}

type KeaCookedMetrics struct {
	CumulativeAssignedAddresses          float64
	DeclinedAddresses                    float64
	Pkt4AckReceived                      float64
	Pkt4AckSent                          float64
	Pkt4DeclineReceived                  float64
	Pkt4DiscoverReceived                 float64
	Pkt4InformReceived                   float64
	Pkt4NakReceived                      float64
	Pkt4NakSent                          float64
	Pkt4OfferReceived                    float64
	Pkt4OfferSent                        float64
	Pkt4ParseFailed                      float64
	Pkt4ReceiveDrop                      float64
	Pkt4Received                         float64
	Pkt4ReleaseReceived                  float64
	Pkt4RequestReceived                  float64
	Pkt4Sent                             float64
	Pkt4UnknownReceived                  float64
	Pkt4LeaseQueryReceived               float64 // pkt4-lease-query-received
	Pkt4LeaseQueryResponseUnknown        float64 // pkt4-lease-query-response-unknown-sent
	Pkt4LeaseQueryResponseUnassignedSent float64 // pkt4-lease-query-response-unassigned-sent
	Pkt4LeaseQueryResponseActiveSent     float64 // pkt4-lease-query-response-active-sent
	V4LeaseReuses                        float64 // v4-lease-reuses
	ReclaimedDeclinedAddresses           float64
	ReclaimedLeases                      float64
	V4AllocationFail                     float64
	V4AllocationFailClasses              float64
	V4AllocationFailNoPools              float64
	V4AllocationFailSharedNetwork        float64
	V4AllocationFailSubnet               float64
	V4ReservationConflicts               float64
	SubnetMetrics                        map[uint64]KeaSubnetMetrics
}

type KeaSubnetMetrics struct {
	Subnet                        string
	SubnetIndex                   uint64
	AssignedAddresses             float64
	CumulativeAssignedAddresses   float64
	DeclinedAddresses             float64
	ReclaimedDeclinedAddresses    float64
	ReclaimedLeases               float64
	TotalAddresses                float64
	V4ReservationConflicts        float64
	V4AllocationFail              float64 // v4-allocation-fail
	V4AllocationFailClasses       float64
	V4AllocationFailNoPools       float64
	V4AllocationFailSharedNetwork float64
	V4AllocationFailSubnet        float64
	V4LeaseReuses                 float64 // v4-lease-reuses
	PoolMetrics                   map[uint64]KeaPoolMetrics
}

type KeaPoolMetrics struct {
	PoolIndex                   uint64
	TotalAddresses              float64
	CumulativeAssignedAddresses float64
	AssignedAddresses           float64
	ReclaimedLeases             float64
	DeclinedAddresses           float64
	ReclaimedDeclinedAddresses  float64
}

func extractCookedMetrics(name string, cooked *KeaCookedMetrics, value float64) {
	switch name {
	case "cumulative-assigned-addresses":
		cooked.CumulativeAssignedAddresses = value
	case "declined-addresses":
		cooked.DeclinedAddresses = value
	case "pkt4-ack-received":
		cooked.Pkt4AckReceived = value
	case "pkt4-ack-sent":
		cooked.Pkt4AckSent = value
	case "pkt4-decline-received":
		cooked.Pkt4DeclineReceived = value
	case "pkt4-discover-received":
		cooked.Pkt4DiscoverReceived = value
	case "pkt4-inform-received":
		cooked.Pkt4InformReceived = value
	case "pkt4-nak-received":
		cooked.Pkt4NakReceived = value
	case "pkt4-nak-sent":
		cooked.Pkt4NakSent = value
	case "pkt4-offer-received":
		cooked.Pkt4OfferReceived = value
	case "pkt4-offer-sent":
		cooked.Pkt4OfferSent = value
	case "pkt4-parse-failed":
		cooked.Pkt4ParseFailed = value
	case "pkt4-receive-drop":
		cooked.Pkt4ReceiveDrop = value
	case "pkt4-received":
		cooked.Pkt4Received = value
	case "pkt4-release-received":
		cooked.Pkt4ReleaseReceived = value
	case "pkt4-request-received":
		cooked.Pkt4RequestReceived = value
	case "pkt4-sent":
		cooked.Pkt4Sent = value
	case "pkt4-unknown-received":
		cooked.Pkt4UnknownReceived = value
	case "reclaimed-declined-addresses":
		cooked.ReclaimedDeclinedAddresses = value
	case "reclaimed-leases":
		cooked.ReclaimedLeases = value
	case "v4-allocation-fail":
		cooked.V4AllocationFail = value
	case "v4-allocation-fail-classes":
		cooked.V4AllocationFailClasses = value
	case "v4-allocation-fail-no-pools":
		cooked.V4AllocationFailNoPools = value
	case "v4-allocation-fail-shared-network":
		cooked.V4AllocationFailSharedNetwork = value
	case "v4-allocation-fail-subnet":
		cooked.V4AllocationFailSubnet = value
	case "v4-reservation-conflicts":
		cooked.V4ReservationConflicts = value
	}
}

func extractSubnetMetric(name string, cooked *KeaCookedMetrics, ml []interface{}) error {
	index, submetric, err := parseMetricNameID(name)
	if err != nil {
		return err
	}
	var snm KeaSubnetMetrics
	if ret, ok := cooked.SubnetMetrics[index]; ok {
		snm = ret
	} else {
		snm = KeaSubnetMetrics{}
	}
	snm.SubnetIndex = index
	val, err := getLatestMetricValue(ml)
	if err != nil {
		return err
	}
	if strings.HasPrefix(submetric, "pool[") {
		err = extractPoolMetric(submetric, &snm, ml)
		if err != nil {
			return err
		}
	} else {
		switch submetric {
		case "assigned-addresses":
			snm.AssignedAddresses = val
		case "cumulative-assigned-addresses":
			snm.CumulativeAssignedAddresses = val
		case "declined-addresses":
			snm.DeclinedAddresses = val
		case "reclaimed-declined-addresses":
			snm.ReclaimedDeclinedAddresses = val
		case "reclaimed-leases":
			snm.ReclaimedLeases = val
		case "total-addresses":
			snm.TotalAddresses = val
		case "v4-reservation-conflicts":
			snm.V4ReservationConflicts = val
		}
	}
	cooked.SubnetMetrics[index] = snm
	return nil
}

func parseMetricNameID(name string) (uint64, string, error) {
	// pool[pid].declined-addresses
	var shortname string
	var index uint64
	openBrkt := strings.Index(name, "[")
	if openBrkt == -1 {
		return index, shortname, fmt.Errorf("could not find opening bracket in submetric name '%s'", name)
	}
	closeBrkt := strings.Index(name, "]")
	if closeBrkt == -1 {
		return index, shortname, fmt.Errorf("could not find closing bracket in submetric name '%s'", name)
	}
	index, err := strconv.ParseUint(name[openBrkt+1:closeBrkt], 10, 64)
	if err != nil {
		return index, shortname, fmt.Errorf("could not parse subnet index from '%s'", name[openBrkt+1:closeBrkt])
	}
	shortname = name[closeBrkt+2:]
	return index, shortname, nil
}

func extractPoolMetric(name string, snm *KeaSubnetMetrics, ml []interface{}) error {
	var pm KeaPoolMetrics
	index, submetric, err := parseMetricNameID(name)
	if err != nil {
		return err
	}
	if ret, ok := snm.PoolMetrics[index]; ok {
		pm = ret
	} else {
		pm = KeaPoolMetrics{}
		pm.PoolIndex = index
	}
	val, err := getLatestMetricValue(ml)
	if err != nil {
		return err
	}
	switch submetric {
	case "total-addresses":
		pm.TotalAddresses = val
	case "cumulative-assigned-addresses":
		pm.CumulativeAssignedAddresses = val
	case "assigned-addresses":
		pm.AssignedAddresses = val
	case "reclaimed-leases":
		pm.ReclaimedLeases = val
	case "declined-addresses":
		pm.DeclinedAddresses = val
	case "reclaimed-declined-addresses":
		pm.ReclaimedDeclinedAddresses = val
	}
	snm.PoolMetrics[pm.PoolIndex] = pm

	return nil
}

func getLatestMetricValue(metricsList []interface{}) (float64, error) {
	ml := make([]metric, 0, len(metricsList))
	for _, metricEntry := range metricsList {
		statPair, ok := metricEntry.([]interface{})
		if !ok {
			return 0, fmt.Errorf("metric entry is not an []interface{}, '%#v'", metricEntry)
		}
		statValue, ok := statPair[0].(float64)
		if !ok {
			return 0, fmt.Errorf("statPair[0] is not a float, '%#v'", statPair[0])
		}
		statTimestamp, ok := statPair[1].(string)
		if !ok {
			return 0, fmt.Errorf("statPair[1] is not a string, '%#v'", statPair[1])
		}
		// "2023-09-14 00:08:10.270215"
		statTime, err := time.Parse("2006-01-02 15:04:05.999999", statTimestamp)
		if err != nil {
			return 0, fmt.Errorf("could not parse time '%s': %w", statTimestamp, err)
		}
		ml = append(ml, metric{statValue, statTime})
	}
	sort.Slice(ml, func(i, j int) bool { return ml[i].t.After(ml[j].t) })
	return ml[0].v, nil
}

type metric struct {
	v float64
	t time.Time
}
