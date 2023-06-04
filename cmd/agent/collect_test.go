package main

// FIXME: need to find a way to implement metrcis collection func, stucked at the moment
// impossible to send pointer to interface implementation receiver

//import (
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/freepaddler/yap-metrics/internal/agent"
//	"github.com/freepaddler/yap-metrics/internal/models"
//)
//

//
//// mock reporter to check only one requested metric
//type singleReporter struct {
//	agent.Reporter
//	RepMetric models.Metrics
//	MType     string
//	MName     string
//}
//
//func newSingleReporter(mType, mName string) *singleReporter {
//	return &singleReporter{
//		MType: mType,
//		MName: mName,
//	}
//}
//
//func (r *singleReporter) Report(m models.Metrics) bool {
//	r.RepMetric = models.Metrics{}
//	if m.Type == r.MType && m.Name == r.MName {
//		r.RepMetric = m
//		return true
//	}
//	return false
//}
//
//func Test_collectMetrics_RandomValue(t *testing.T) {
//	sc := agent.NewStatsCollector()
//	var r agent.Reporter
//	sr := singleReporter{
//		MType: models.Gauge,
//		MName: "RandomValue",
//	}
//	r = &sr
//	collectMetrics(sc)
//	sc.Report(r)
//	rand := sr.RepMetric.Gauge
//	assert.Equal(t, models.Gauge, sr.RepMetric.Type)
//	assert.Equal(t, "RandomValue", sr.RepMetric.Name)
//	collectMetrics(sc)
//	sc.Report(r)
//	assert.NotEqual(t, rand, sr.RepMetric.Gauge)
//}
//
