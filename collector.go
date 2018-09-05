// collector.go - implement the radosgw service exporter

package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/oshynsong/radosgw_exporter/radosgw"
)

const radosgwNamespace = "radosgw"

type RadosgwCollector struct {
	client *radosgw.Client

	// BytesSent shows the total send throughput.
	bytesSent []prometheus.Gauge

	// BytesRecv shows the total received throughput.
	bytesRecv []prometheus.Gauge

	// Ops shows the total operation called times.
	ops []prometheus.Gauge

	// OpsOK shows the total operation called times successfully.
	opsOK []prometheus.Gauge

	// numObjects shows the total object number.
	numObjects []prometheus.Gauge

	// capacity shows the current disk space occupied by all objects.
	capacity []prometheus.Gauge
}

func NewRadosgwCollector(endpoint, ak, sk string) (*RadosgwCollector, error) {
	cli, err := radosgw.NewClient(endpoint, ak, sk)
	if err != nil {
		return nil, err
	}
	return &RadosgwCollector{client: cli}, nil
}

func (r *RadosgwCollector) Describe(ch chan<- *prometheus.Desc) {
	r.collecting()
	gauges := r.allGauges()
	for i := range gauges {
		ch <- gauges[i].Desc()
	}
}

func (r *RadosgwCollector) Collect(ch chan<- prometheus.Metric) {
	r.collecting()
	gauges := r.allGauges()
	for i := range gauges {
		ch <- gauges[i]
	}
}

func (r *RadosgwCollector) allGauges() []prometheus.Gauge {
	result := make([]prometheus.Gauge, 0)
	result = append(result, r.bytesSent...)
	result = append(result, r.bytesRecv...)
	result = append(result, r.ops...)
	result = append(result, r.opsOK...)
	result = append(result, r.numObjects...)
	result = append(result, r.capacity...)
	return result
}

func (r *RadosgwCollector) collecting() {
	// Collect the bucket usage data
	status, bucketStats, err := r.client.GetBucket("", "", true)
	if err != nil || status > 200 {
		fmt.Printf("collect the radosgw bucket stats failed: %v", err)
		return
	}
	r.numObjects = make([]prometheus.Gauge, 0)
	r.capacity = make([]prometheus.Gauge, 0)
	for i := range bucketStats {
		stats := bucketStats[i].Stats
		numObjectsGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: radosgwNamespace,
			Name:      "num_objects",
			Help:      "total object number",
			ConstLabels: prometheus.Labels{
				"user":   stats.Owner,
				"bucket": stats.Bucket,
			},
		})
		numObjectsGauge.Set(float64(stats.Usage.RgwMain.NumObjects))
		r.numObjects = append(r.numObjects, numObjectsGauge)

		capacityGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: radosgwNamespace,
			Name:      "capacity",
			Help:      "current disk space usage of all objects",
			ConstLabels: prometheus.Labels{
				"user":   stats.Owner,
				"bucket": stats.Bucket,
			},
		})
		capacityGauge.Set(float64(stats.Usage.RgwMain.Size))
		r.capacity = append(r.capacity, capacityGauge)
	}

	// Collect the API usage data by users
	r.bytesSent = make([]prometheus.Gauge, 0)
	r.bytesRecv = make([]prometheus.Gauge, 0)
	r.ops = make([]prometheus.Gauge, 0)
	r.opsOK = make([]prometheus.Gauge, 0)
	status, usage, err := r.client.GetUsage("", nil, nil, false, false)
	if err != nil || status > 200 {
		fmt.Printf("collect the radosgw usage metrics failed: %v", err)
		return
	}
	if len(usage.Entries) == 0 {
		return
	}
	for i := range usage.Entries {
		userLabel := usage.Entries[i].User
		buckets := usage.Entries[i].Buckets
		for k := range buckets {
			bucketLabel := buckets[k].Bucket
			categories := buckets[k].Categories
			for c := range categories {
				apiLabel := categories[c].Category
				bytesSentGauge := prometheus.NewGauge(prometheus.GaugeOpts{
					Namespace: radosgwNamespace,
					Name:      "bytes_sent_total",
					Help:      "currently total sent throughput",
					ConstLabels: prometheus.Labels{
						"user":   userLabel,
						"bucket": bucketLabel,
						"api":    apiLabel,
					},
				})
				bytesSentGauge.Set(float64(categories[c].BytesSent))
				r.bytesSent = append(r.bytesSent, bytesSentGauge)

				bytesRecvGauge := prometheus.NewGauge(prometheus.GaugeOpts{
					Namespace: radosgwNamespace,
					Name:      "bytes_recv_total",
					Help:      "currently total recv throughput",
					ConstLabels: prometheus.Labels{
						"user":   userLabel,
						"bucket": bucketLabel,
						"api":    apiLabel,
					},
				})
				bytesRecvGauge.Set(float64(categories[c].BytesReceived))
				r.bytesRecv = append(r.bytesRecv, bytesRecvGauge)

				opsGauge := prometheus.NewGauge(prometheus.GaugeOpts{
					Namespace: radosgwNamespace,
					Name:      "ops_total",
					Help:      "currently total ops",
					ConstLabels: prometheus.Labels{
						"user":   userLabel,
						"bucket": bucketLabel,
						"api":    apiLabel,
					},
				})
				opsGauge.Set(float64(categories[c].Ops))
				r.ops = append(r.ops, opsGauge)

				opsOKGauge := prometheus.NewGauge(prometheus.GaugeOpts{
					Namespace: radosgwNamespace,
					Name:      "ops_ok_total",
					Help:      "currently total ops ok",
					ConstLabels: prometheus.Labels{
						"user":   userLabel,
						"bucket": bucketLabel,
						"api":    apiLabel,
					},
				})
				opsOKGauge.Set(float64(categories[c].SuccessfulOps))
				r.opsOK = append(r.opsOK, opsOKGauge)
			}
		}
	}
}
