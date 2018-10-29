// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mgomonitor_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"gopkg.in/mgo.v2"

	"github.com/juju/mgomonitor"
)

var _ prometheus.Collector = (*mgomonitor.Collector)(nil)

var (
	clusterDesc      = prometheus.NewDesc("test_mgo_clusters", "Number of alive clusters.", nil, nil)
	masterConnDesc   = prometheus.NewDesc("test_mgo_master_connections", "Number of master connections.", nil, nil)
	slaveConnDesc    = prometheus.NewDesc("test_mgo_slave_connections", "Number of slave connections.", nil, nil)
	sentOpsDesc      = prometheus.NewDesc("test_mgo_sent_operations", "Number of operations sent.", nil, nil)
	receivedOpsDesc  = prometheus.NewDesc("test_mgo_received_operations", "Number of operations received.", nil, nil)
	receivedDocsDesc = prometheus.NewDesc("test_mgo_received_documents", "Number of documents received.", nil, nil)
	socketsAliveDesc = prometheus.NewDesc("test_mgo_sockets_alive", "Number of alive sockets.", nil, nil)
	socketsInUseDesc = prometheus.NewDesc("test_mgo_sockets_in_use", "Number of in use sockets.", nil, nil)
	socketRefsDesc   = prometheus.NewDesc("test_mgo_socket_references", "Number of references to sockets.", nil, nil)
)

func TestDescribe(t *testing.T) {
	c := qt.New(t)

	expectDescriptions := []*prometheus.Desc{
		clusterDesc,
		masterConnDesc,
		slaveConnDesc,
		sentOpsDesc,
		receivedOpsDesc,
		receivedDocsDesc,
		socketsAliveDesc,
		socketsInUseDesc,
		socketRefsDesc,
	}

	coll := mgomonitor.NewCollector("test")
	ch := make(chan *prometheus.Desc)
	go func() {
		defer close(ch)
		coll.Describe(ch)
	}()

	var obtainedDescriptions []*prometheus.Desc
	for d := range ch {
		obtainedDescriptions = append(obtainedDescriptions, d)
	}
	c.Assert(obtainedDescriptions, deepEquals, expectDescriptions)
}

var deepEquals = qt.CmpEquals(cmp.AllowUnexported(prometheus.Desc{}))

func TestCollect(t *testing.T) {
	c := qt.New(t)

	coll := mgomonitor.NewCollector("test")
	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		coll.Collect(ch)
	}()
	stats := mgo.GetStats()
	var obtainedMetrics []prometheus.Metric
	for m := range ch {
		obtainedMetrics = append(obtainedMetrics, m)
	}
	c.Assert(obtainedMetrics, qt.HasLen, 9)

	c.Assert(obtainedMetrics[0].Desc(), deepEquals, clusterDesc)
	assertGauge(c, obtainedMetrics[0], float64(stats.Clusters))

	c.Assert(obtainedMetrics[1].Desc(), deepEquals, masterConnDesc)
	assertGauge(c, obtainedMetrics[1], float64(stats.MasterConns))

	c.Assert(obtainedMetrics[2].Desc(), deepEquals, slaveConnDesc)
	assertGauge(c, obtainedMetrics[2], float64(stats.SlaveConns))

	c.Assert(obtainedMetrics[3].Desc(), deepEquals, sentOpsDesc)
	assertGauge(c, obtainedMetrics[3], float64(stats.SentOps))

	c.Assert(obtainedMetrics[4].Desc(), deepEquals, receivedOpsDesc)
	assertGauge(c, obtainedMetrics[4], float64(stats.ReceivedOps))

	c.Assert(obtainedMetrics[5].Desc(), deepEquals, receivedDocsDesc)
	assertGauge(c, obtainedMetrics[5], float64(stats.ReceivedDocs))

	c.Assert(obtainedMetrics[6].Desc(), deepEquals, socketsAliveDesc)
	assertGauge(c, obtainedMetrics[6], float64(stats.SocketsAlive))

	c.Assert(obtainedMetrics[7].Desc(), deepEquals, socketsInUseDesc)
	assertGauge(c, obtainedMetrics[7], float64(stats.SocketsInUse))

	c.Assert(obtainedMetrics[8].Desc(), deepEquals, socketRefsDesc)
	assertGauge(c, obtainedMetrics[8], float64(stats.SocketRefs))
}

func assertGauge(c *qt.C, m prometheus.Metric, value float64) {
	var metric dto.Metric
	err := m.Write(&metric)
	c.Assert(err, qt.Equals, nil)
	c.Assert(metric.Gauge, qt.Not(qt.IsNil))
	c.Assert(*metric.Gauge.Value, qt.Equals, value)
}
