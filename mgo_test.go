// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package monitoring_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	gc "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"

	"github.com/cloud-green/monitoring"
)

var _ prometheus.Collector = (*monitoring.MgoStatsCollector)(nil)

type mgoStatsSuite struct {
	testing.MgoSuite
}

var _ = gc.Suite(&mgoStatsSuite{})

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

func (s *mgoStatsSuite) TestDescribe(c *gc.C) {
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

	coll := monitoring.NewMgoStatsCollector("test")
	ch := make(chan *prometheus.Desc)
	go func() {
		defer close(ch)
		coll.Describe(ch)
	}()

	var obtainedDescriptions []*prometheus.Desc
	for d := range ch {
		obtainedDescriptions = append(obtainedDescriptions, d)
	}
	c.Assert(obtainedDescriptions, jc.DeepEquals, expectDescriptions)
}

func (s *mgoStatsSuite) TestCollect(c *gc.C) {
	coll := monitoring.NewMgoStatsCollector("test")
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
	c.Assert(obtainedMetrics, gc.HasLen, 9)

	c.Assert(obtainedMetrics[0].Desc(), jc.DeepEquals, clusterDesc)
	s.assertGauge(c, obtainedMetrics[0], float64(stats.Clusters))

	c.Assert(obtainedMetrics[1].Desc(), jc.DeepEquals, masterConnDesc)
	s.assertGauge(c, obtainedMetrics[1], float64(stats.MasterConns))

	c.Assert(obtainedMetrics[2].Desc(), jc.DeepEquals, slaveConnDesc)
	s.assertGauge(c, obtainedMetrics[2], float64(stats.SlaveConns))

	c.Assert(obtainedMetrics[3].Desc(), jc.DeepEquals, sentOpsDesc)
	s.assertGauge(c, obtainedMetrics[3], float64(stats.SentOps))

	c.Assert(obtainedMetrics[4].Desc(), jc.DeepEquals, receivedOpsDesc)
	s.assertGauge(c, obtainedMetrics[4], float64(stats.ReceivedOps))

	c.Assert(obtainedMetrics[5].Desc(), jc.DeepEquals, receivedDocsDesc)
	s.assertGauge(c, obtainedMetrics[5], float64(stats.ReceivedDocs))

	c.Assert(obtainedMetrics[6].Desc(), jc.DeepEquals, socketsAliveDesc)
	s.assertGauge(c, obtainedMetrics[6], float64(stats.SocketsAlive))

	c.Assert(obtainedMetrics[7].Desc(), jc.DeepEquals, socketsInUseDesc)
	s.assertGauge(c, obtainedMetrics[7], float64(stats.SocketsInUse))

	c.Assert(obtainedMetrics[8].Desc(), jc.DeepEquals, socketRefsDesc)
	s.assertGauge(c, obtainedMetrics[8], float64(stats.SocketRefs))
}

func (s *mgoStatsSuite) assertGauge(c *gc.C, m prometheus.Metric, value float64) {
	var metric dto.Metric
	err := m.Write(&metric)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(metric.Gauge, gc.Not(gc.IsNil))
	c.Assert(*metric.Gauge.Value, gc.Equals, value)
}
