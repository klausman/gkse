package main

import (
	"flag"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var namespace = flag.String("namespace", "kea", "Namespace (prefix) to use for Prometheus metrics")

func newKeaCollector(namespace string) prometheus.Collector {
	subnetlabels := []string{"subnetidx", "subnet"}
	poollabels := make([]string, 0, len(subnetlabels))
	copy(subnetlabels, poollabels)
	poollabels = append(poollabels, "poolidx")

	c4 := jsonCollector4{
		namespace:                   namespace,
		CumulativeAssignedAddresses: prometheus.NewDesc(namespace+"_addresses_assigned_total", "Cumulative number of addresses that have been assigned since server startup", nil, nil),
		DeclinedAddresses:           prometheus.NewDesc(namespace+"_addresses_declined_total", "Number of IPv4 addresses that are currently declined; a count of the number of leases currently unavailable", nil, nil),
		// Totals (v4)
		Pkt4Received: prometheus.NewDesc(namespace+"_v4_packets_received_total", "Number of DHCPv4 packets received. This includes all packets: valid, bogus, corrupted, rejected, etc.", nil, nil),
		Pkt4Sent:     prometheus.NewDesc(namespace+"_v4_packets_sent_total", "Number of DHCPv4 packets sent", nil, nil),
		// RX types (v4)
		Pkt4AckReceived:      prometheus.NewDesc(namespace+"_v4_packet_types_received_total", "Number v4 of packets received", nil, map[string]string{"pkttype": "ack"}),
		Pkt4DeclineReceived:  prometheus.NewDesc(namespace+"_v4_packet_types_received_total", "Number v4 of packets received", nil, map[string]string{"pkttype": "decline"}),
		Pkt4DiscoverReceived: prometheus.NewDesc(namespace+"_v4_packet_types_received_total", "Number v4 of packets received", nil, map[string]string{"pkttype": "discover"}),
		Pkt4InformReceived:   prometheus.NewDesc(namespace+"_v4_packet_types_received_total", "Number v4 of packets received", nil, map[string]string{"pkttype": "inform"}),
		Pkt4NakReceived:      prometheus.NewDesc(namespace+"_v4_packet_types_received_total", "Number v4 of packets received", nil, map[string]string{"pkttype": "nak"}),
		Pkt4OfferReceived:    prometheus.NewDesc(namespace+"_v4_packet_types_received_total", "Number v4 of packets received", nil, map[string]string{"pkttype": "offer"}),
		Pkt4ReleaseReceived:  prometheus.NewDesc(namespace+"_v4_packet_types_received_total", "Number v4 of packets received", nil, map[string]string{"pkttype": "release"}),
		Pkt4RequestReceived:  prometheus.NewDesc(namespace+"_v4_packet_types_received_total", "Number v4 of packets received", nil, map[string]string{"pkttype": "request"}),
		Pkt4UnknownReceived:  prometheus.NewDesc(namespace+"_v4_packet_types_received_total", "Number v4 of packets received", nil, map[string]string{"pkttype": "unknown"}),
		// TX types (v4)
		Pkt4AckSent:   prometheus.NewDesc(namespace+"_v4_packet_types_sent_total", "Number of v4 packets sent", nil, map[string]string{"pkttype": "ack"}),
		Pkt4NakSent:   prometheus.NewDesc(namespace+"_v4_packet_types_sent_total", "Number of v4 packets sent", nil, map[string]string{"pkttype": "nak"}),
		Pkt4OfferSent: prometheus.NewDesc(namespace+"_v4_packet_types_sent_total", "Number of v4 packets sent", nil, map[string]string{"pkttype": "offer"}),
		// Misc (v4)
		Pkt4ParseFailed:               prometheus.NewDesc(namespace+"_v4_packets_parse_failed_total", "Number of incoming packets that could not be parsed", nil, nil),
		Pkt4ReceiveDrop:               prometheus.NewDesc(namespace+"_v4_packets_dropped_on_receive_total", "Number of incoming packets that were dropped", nil, nil),
		V4AllocationFailClasses:       prometheus.NewDesc(namespace+"_v4_allocation_failures_classes_total", "Number of address allocation failures when the client's packet belongs to one or more classes", nil, nil),
		V4AllocationFailNoPools:       prometheus.NewDesc(namespace+"_v4_allocation_failures_no_pools_total", "Number of address allocation failures because the server could not use any configured pools for a particular client", nil, nil),
		V4AllocationFail:              prometheus.NewDesc(namespace+"_v4_allocation_failures_total", "Number of total address allocation failures", nil, nil),
		V4AllocationFailSharedNetwork: prometheus.NewDesc(namespace+"_v4_allocation_failures_shared_network_total", "Number of address allocation", nil, nil),
		V4AllocationFailSubnet:        prometheus.NewDesc(namespace+"_v4_allocation_failures_subnet_total", "Number of address allocation failures for a particular client connected to a subnet that does not belong to a shared network", nil, nil),
		V4ReservationConflicts:        prometheus.NewDesc(namespace+"_v4_reservation_conflicts_total", "Number of host reservation allocation conflicts which have occurred across every subnet", nil, nil),
		// Misc
		ReclaimedDeclinedAddresses: prometheus.NewDesc(namespace+"_reclaimed_declined_addresses_total", "Number of IPv4 addresses that were declined, but have now been recovered", nil, nil),
		ReclaimedLeases:            prometheus.NewDesc(namespace+"_reclaimed_leases_total", "Number of expired leases that have been reclaimed since server startup", nil, nil),
		// Subnet metrics
		SubnetAssignedAddresses:               prometheus.NewDesc(namespace+"_subnet_assigned_addresses", "Number of assigned addresses in a given subnet", subnetlabels, nil),
		SubnetAssignedAddressesTotal:          prometheus.NewDesc(namespace+"_subnet_assigned_addresses_total", "Cumulative number of assigned addresses in a given subnet", subnetlabels, nil),
		SubnetDeclinedAddressesTotal:          prometheus.NewDesc(namespace+"_subnet_declined_addresses_total", "Number of IPv4 addresses that are currently declined in a given subnet; a count of the number of leases currently unavailable", subnetlabels, nil),
		SubnetReclaimedDeclinedAddressesTotal: prometheus.NewDesc(namespace+"_subnet_reclaimed_declined_addresses", "Number of IPv4 addresses that were declined, but have now been recovered", subnetlabels, nil),
		SubnetReclaimedLeasesTotal:            prometheus.NewDesc(namespace+"_subnet_reclaimed_leases_total", "Number of expired leases associated with a given subnet that have been reclaimed since server startup", subnetlabels, nil),
		SubnetAddressesTotal:                  prometheus.NewDesc(namespace+"_subnet_addresses", "Total number of addresses available for DHCPv4 management for a given subnet; in other words, this is the count of all addresses in all configured pools", subnetlabels, nil),
		SubnetReservationConflictsTotal:       prometheus.NewDesc(namespace+"_subnet_reservation_conflicts_total", "Number of host reservation allocation conflicts which have occurred in a specific subnet.", subnetlabels, nil),
		// Pool metrics
		PoolTotalAddresses:              prometheus.NewDesc(namespace+"_subnet_pool_addresses", "Total number of addresses available for DHCPv4 management for a given subnet pool", poollabels, nil),
		PoolCumulativeAssignedAddresses: prometheus.NewDesc(namespace+"_subnet_pool_addresses_assigned_total", "Cumulative number of assigned addresses in a given subnet pool", poollabels, nil),
		PoolAssignedAddresses:           prometheus.NewDesc(namespace+"_subnet_pool_assigned_addresses", "Number of assigned addresses in a given subnet pool", poollabels, nil),
		PoolReclaimedLeases:             prometheus.NewDesc(namespace+"_subnet_pool_reclaimed_leases_total", "Number of expired leases associated with a given subnet pool that have been reclaimed since server startup", poollabels, nil),
		PoolDeclinedAddresses:           prometheus.NewDesc(namespace+"_subnet_pool_addresses_declined_total", "Number of IPv4 addresses that are currently declined in a given subnet pool; a count of the number of leases currently unavailable", poollabels, nil),
		PoolReclaimedDeclinedAddresses:  prometheus.NewDesc(namespace+"_subnet_pool_reclaimed_declined_addresses_total", "Number of IPv4 addresses that were declined, but have now been recovered in this pool", poollabels, nil),
	}
	return &c4
}

type jsonCollector4 struct {
	namespace                   string
	CumulativeAssignedAddresses *prometheus.Desc
	DeclinedAddresses           *prometheus.Desc
	// Totals
	Pkt4Received *prometheus.Desc
	Pkt4Sent     *prometheus.Desc
	// RX types
	Pkt4AckReceived      *prometheus.Desc
	Pkt4DeclineReceived  *prometheus.Desc
	Pkt4DiscoverReceived *prometheus.Desc
	Pkt4InformReceived   *prometheus.Desc
	Pkt4NakReceived      *prometheus.Desc
	Pkt4OfferReceived    *prometheus.Desc
	Pkt4ReleaseReceived  *prometheus.Desc
	Pkt4RequestReceived  *prometheus.Desc
	Pkt4UnknownReceived  *prometheus.Desc
	// TX types
	Pkt4AckSent   *prometheus.Desc
	Pkt4NakSent   *prometheus.Desc
	Pkt4OfferSent *prometheus.Desc
	// Misc
	Pkt4ParseFailed               *prometheus.Desc
	Pkt4ReceiveDrop               *prometheus.Desc
	ReclaimedDeclinedAddresses    *prometheus.Desc
	ReclaimedLeases               *prometheus.Desc
	V4AllocationFailClasses       *prometheus.Desc
	V4AllocationFailNoPools       *prometheus.Desc
	V4AllocationFail              *prometheus.Desc
	V4AllocationFailSharedNetwork *prometheus.Desc
	V4AllocationFailSubnet        *prometheus.Desc
	V4ReservationConflicts        *prometheus.Desc
	// Subnet metrics
	SubnetAssignedAddresses               *prometheus.Desc
	SubnetAssignedAddressesTotal          *prometheus.Desc
	SubnetDeclinedAddressesTotal          *prometheus.Desc
	SubnetReclaimedDeclinedAddressesTotal *prometheus.Desc
	SubnetReclaimedLeasesTotal            *prometheus.Desc
	SubnetAddressesTotal                  *prometheus.Desc
	SubnetReservationConflictsTotal       *prometheus.Desc
	// Pool metrics
	PoolTotalAddresses              *prometheus.Desc
	PoolCumulativeAssignedAddresses *prometheus.Desc
	PoolAssignedAddresses           *prometheus.Desc
	PoolReclaimedLeases             *prometheus.Desc
	PoolDeclinedAddresses           *prometheus.Desc
	PoolReclaimedDeclinedAddresses  *prometheus.Desc
}

func (c *jsonCollector4) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *jsonCollector4) Collect(ch chan<- prometheus.Metric) {
	var rawJSON []byte
	var err error
	logger.Debug("Fetching stats from Kea")
	rawJSON, err = getStatsJSON()
	if err != nil {
		logger.Error("Could not fetch stats from Kea", "error", err)
		return
	}
	cooked, err := parseStats(rawJSON)
	if err != nil {
		logger.Error("Could not parse raw JSON stats", "error", err)
		return
	}
	config, err := queryConfig()
	if err != nil {
		logger.Error("Could not query Kea config", "error", err)
		return
	}
	logger.Debug("Sending stats to channel")
	ch <- prometheus.MustNewConstMetric(
		c.CumulativeAssignedAddresses, prometheus.CounterValue, cooked.CumulativeAssignedAddresses)
	ch <- prometheus.MustNewConstMetric(
		c.DeclinedAddresses, prometheus.GaugeValue, cooked.DeclinedAddresses)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4AckReceived, prometheus.CounterValue, cooked.Pkt4AckReceived)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4AckSent, prometheus.CounterValue, cooked.Pkt4AckSent)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4DeclineReceived, prometheus.CounterValue, cooked.Pkt4DeclineReceived)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4DiscoverReceived, prometheus.CounterValue, cooked.Pkt4DiscoverReceived)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4InformReceived, prometheus.CounterValue, cooked.Pkt4InformReceived)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4NakReceived, prometheus.CounterValue, cooked.Pkt4NakReceived)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4NakSent, prometheus.CounterValue, cooked.Pkt4NakSent)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4OfferReceived, prometheus.CounterValue, cooked.Pkt4OfferReceived)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4OfferSent, prometheus.CounterValue, cooked.Pkt4OfferSent)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4ParseFailed, prometheus.CounterValue, cooked.Pkt4ParseFailed)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4ReceiveDrop, prometheus.CounterValue, cooked.Pkt4ReceiveDrop)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4Received, prometheus.CounterValue, cooked.Pkt4Received)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4ReleaseReceived, prometheus.CounterValue, cooked.Pkt4ReleaseReceived)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4RequestReceived, prometheus.CounterValue, cooked.Pkt4RequestReceived)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4Sent, prometheus.CounterValue, cooked.Pkt4Sent)
	ch <- prometheus.MustNewConstMetric(
		c.Pkt4UnknownReceived, prometheus.CounterValue, cooked.Pkt4UnknownReceived)
	ch <- prometheus.MustNewConstMetric(
		c.ReclaimedDeclinedAddresses, prometheus.CounterValue, cooked.ReclaimedDeclinedAddresses)
	ch <- prometheus.MustNewConstMetric(
		c.ReclaimedLeases, prometheus.CounterValue, cooked.ReclaimedLeases)
	ch <- prometheus.MustNewConstMetric(
		c.V4AllocationFail, prometheus.CounterValue, cooked.V4AllocationFail)
	ch <- prometheus.MustNewConstMetric(
		c.V4AllocationFailClasses, prometheus.CounterValue, cooked.V4AllocationFailClasses)
	ch <- prometheus.MustNewConstMetric(
		c.V4AllocationFailNoPools, prometheus.CounterValue, cooked.V4AllocationFailNoPools)
	ch <- prometheus.MustNewConstMetric(
		c.V4AllocationFailSharedNetwork, prometheus.CounterValue, cooked.V4AllocationFailSharedNetwork)
	ch <- prometheus.MustNewConstMetric(
		c.V4AllocationFailSubnet, prometheus.CounterValue, cooked.V4AllocationFailSubnet)
	ch <- prometheus.MustNewConstMetric(
		c.V4ReservationConflicts, prometheus.CounterValue, cooked.V4ReservationConflicts)
	for _, subnetMetrics := range cooked.SubnetMetrics {
		subnetvalues := []string{fmt.Sprintf("%d", subnetMetrics.SubnetIndex)}
		sn, err := config.subnetFromID(4, subnetMetrics.SubnetIndex)
		if err != nil {
			logger.Error("v4 Subnet of index has no entry in the config", "subnetIndex", subnetMetrics.SubnetIndex)
			sn = "unknown"
		}
		subnetvalues = append(subnetvalues, sn)
		ch <- prometheus.MustNewConstMetric(c.SubnetAssignedAddresses,
			prometheus.GaugeValue, subnetMetrics.AssignedAddresses, subnetvalues...)
		ch <- prometheus.MustNewConstMetric(c.SubnetAssignedAddressesTotal,
			prometheus.CounterValue, subnetMetrics.CumulativeAssignedAddresses, subnetvalues...)
		ch <- prometheus.MustNewConstMetric(c.SubnetDeclinedAddressesTotal,
			prometheus.GaugeValue, subnetMetrics.DeclinedAddresses, subnetvalues...)
		ch <- prometheus.MustNewConstMetric(c.SubnetReclaimedDeclinedAddressesTotal,
			prometheus.CounterValue, subnetMetrics.ReclaimedDeclinedAddresses, subnetvalues...)
		ch <- prometheus.MustNewConstMetric(c.SubnetReclaimedLeasesTotal,
			prometheus.CounterValue, subnetMetrics.ReclaimedLeases, subnetvalues...)
		ch <- prometheus.MustNewConstMetric(c.SubnetAddressesTotal,
			prometheus.GaugeValue, subnetMetrics.TotalAddresses, subnetvalues...)
		ch <- prometheus.MustNewConstMetric(c.SubnetReservationConflictsTotal,
			prometheus.CounterValue, subnetMetrics.V4ReservationConflicts, subnetvalues...)
		for _, poolMetrics := range subnetMetrics.PoolMetrics {
			poolValues := []string{fmt.Sprintf("%d", poolMetrics.PoolIndex)}
			pn, err := config.subnetFromID(4, poolMetrics.PoolIndex)
			if err != nil {
				logger.Error("v4 Subnet in Pool has no entry in the config", "subnetIndex", subnetMetrics.SubnetIndex, "poolIndex", poolMetrics.PoolIndex)
				pn = "unknown"
			}
			poolValues = append(poolValues, pn)
			ch <- prometheus.MustNewConstMetric(c.PoolTotalAddresses,
				prometheus.GaugeValue, poolMetrics.TotalAddresses, poolValues...)
			ch <- prometheus.MustNewConstMetric(c.PoolCumulativeAssignedAddresses,
				prometheus.CounterValue, poolMetrics.CumulativeAssignedAddresses, poolValues...)
			ch <- prometheus.MustNewConstMetric(c.PoolAssignedAddresses,
				prometheus.GaugeValue, poolMetrics.AssignedAddresses, poolValues...)
			ch <- prometheus.MustNewConstMetric(c.PoolReclaimedLeases,
				prometheus.CounterValue, poolMetrics.ReclaimedLeases, poolValues...)
			ch <- prometheus.MustNewConstMetric(c.PoolDeclinedAddresses,
				prometheus.GaugeValue, poolMetrics.DeclinedAddresses, poolValues...)
			ch <- prometheus.MustNewConstMetric(c.PoolReclaimedDeclinedAddresses,
				prometheus.GaugeValue, poolMetrics.ReclaimedDeclinedAddresses, poolValues...)
		}
	}
	logger.Debug("Sending stats to channel complete")
}
