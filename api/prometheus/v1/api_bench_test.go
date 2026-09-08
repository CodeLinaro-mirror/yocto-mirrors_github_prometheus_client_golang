// Copyright 2019 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package v1

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/prometheus/common/model"
)

func generateData(timeseries, datapoints int) (floatMatrix, histogramMatrix model.Matrix) {
	for i := 0; i < timeseries; i++ {
		lset := map[model.LabelName]model.LabelValue{
			model.MetricNameLabel: model.LabelValue("timeseries_" + strconv.Itoa(i)),
			"foo":                 "bar",
		}
		now := model.Time(1677587274055)
		floats := make([]model.SamplePair, datapoints)
		histograms := make([]model.SampleHistogramPair, datapoints)

		for x := datapoints; x > 0; x-- {
			f := float64(x)
			floats[x-1] = model.SamplePair{
				// Set the time back assuming a 15s interval. Since this is used for
				// Marshal/Unmarshal testing the actual interval doesn't matter.
				Timestamp: now.Add(time.Second * -15 * time.Duration(x)),
				Value:     model.SampleValue(f),
			}
			histograms[x-1] = model.SampleHistogramPair{
				Timestamp: now.Add(time.Second * -15 * time.Duration(x)),
				Histogram: &model.SampleHistogram{
					Count: model.FloatString(13.5 * f),
					Sum:   model.FloatString(.1 * f),
					Buckets: model.HistogramBuckets{
						{
							Boundaries: 1,
							Lower:      -4870.992343051145,
							Upper:      -4466.7196729968955,
							Count:      model.FloatString(1 * f),
						},
						{
							Boundaries: 1,
							Lower:      -861.0779292198035,
							Upper:      -789.6119426088657,
							Count:      model.FloatString(2 * f),
						},
						{
							Boundaries: 1,
							Lower:      -558.3399591246119,
							Upper:      -512,
							Count:      model.FloatString(3 * f),
						},
						{
							Boundaries: 0,
							Lower:      2048,
							Upper:      2233.3598364984477,
							Count:      model.FloatString(1.5 * f),
						},
						{
							Boundaries: 0,
							Lower:      2896.3093757400984,
							Upper:      3158.4477704354626,
							Count:      model.FloatString(2.5 * f),
						},
						{
							Boundaries: 0,
							Lower:      4466.7196729968955,
							Upper:      4870.992343051145,
							Count:      model.FloatString(3.5 * f),
						},
					},
				},
			}
		}

		fss := &model.SampleStream{
			Metric: model.Metric(lset),
			Values: floats,
		}
		hss := &model.SampleStream{
			Metric:     model.Metric(lset),
			Histograms: histograms,
		}

		floatMatrix = append(floatMatrix, fss)
		histogramMatrix = append(histogramMatrix, hss)
	}
	return floatMatrix, histogramMatrix
}

func BenchmarkSamplesJsonSerialization(b *testing.B) {
	for _, timeseriesCount := range []int{10, 100, 1000} {
		b.Run(strconv.Itoa(timeseriesCount), func(b *testing.B) {
			for _, datapointCount := range []int{10, 100, 1000} {
				b.Run(strconv.Itoa(datapointCount), func(b *testing.B) {
					floats, histograms := generateData(timeseriesCount, datapointCount)

					floatBytes, err := json.Marshal(floats)
					if err != nil {
						b.Fatalf("Error marshaling: %v", err)
					}
					histogramBytes, err := json.Marshal(histograms)
					if err != nil {
						b.Fatalf("Error marshaling: %v", err)
					}

					b.Run("marshal", func(b *testing.B) {
						b.Run("encoding/json/floats", func(b *testing.B) {
							b.ReportAllocs()
							for i := 0; i < b.N; i++ {
								if _, err := json.Marshal(floats); err != nil {
									b.Fatal(err)
								}
							}
						})
						b.Run("jsoniter/floats", func(b *testing.B) {
							b.ReportAllocs()
							for i := 0; i < b.N; i++ {
								if _, err := jsoniter.Marshal(floats); err != nil {
									b.Fatal(err)
								}
							}
						})
						b.Run("encoding/json/histograms", func(b *testing.B) {
							b.ReportAllocs()
							for i := 0; i < b.N; i++ {
								if _, err := json.Marshal(histograms); err != nil {
									b.Fatal(err)
								}
							}
						})
						b.Run("jsoniter/histograms", func(b *testing.B) {
							b.ReportAllocs()
							for i := 0; i < b.N; i++ {
								if _, err := jsoniter.Marshal(histograms); err != nil {
									b.Fatal(err)
								}
							}
						})
					})

					b.Run("unmarshal", func(b *testing.B) {
						b.Run("encoding/json/floats", func(b *testing.B) {
							b.ReportAllocs()
							var m model.Matrix
							for i := 0; i < b.N; i++ {
								if err := json.Unmarshal(floatBytes, &m); err != nil {
									b.Fatal(err)
								}
							}
						})
						b.Run("jsoniter/floats", func(b *testing.B) {
							b.ReportAllocs()
							var m model.Matrix
							for i := 0; i < b.N; i++ {
								if err := jsoniter.Unmarshal(floatBytes, &m); err != nil {
									b.Fatal(err)
								}
							}
						})
						b.Run("encoding/json/histograms", func(b *testing.B) {
							b.ReportAllocs()
							var m model.Matrix
							for i := 0; i < b.N; i++ {
								if err := json.Unmarshal(histogramBytes, &m); err != nil {
									b.Fatal(err)
								}
							}
						})
						b.Run("jsoniter/histograms", func(b *testing.B) {
							b.ReportAllocs()
							var m model.Matrix
							for i := 0; i < b.N; i++ {
								if err := jsoniter.Unmarshal(histogramBytes, &m); err != nil {
									b.Fatal(err)
								}
							}
						})
					})
				})
			}
		})
	}
}

func BenchmarkRuleGroup(b *testing.B) {
	alertingRuleJSON, err := json.Marshal(struct {
		Type         RuleType `json:"type"`
		AlertingRule `json:""`
	}{
		Type: RuleTypeAlerting,
		AlertingRule: AlertingRule{
			Name:        "HighRequestLatency",
			Query:       "job:request_latency_seconds:mean5m{job=\"myjob\"} > 0.5",
			Duration:    600,
			Labels:      model.LabelSet{"severity": "page"},
			Annotations: model.LabelSet{"summary": "High request latency"},
			Alerts: []*Alert{{
				ActiveAt:    time.Now().UTC(),
				Annotations: model.LabelSet{"summary": "High request latency"},
				Labels:      model.LabelSet{"alertname": "HighRequestLatency", "severity": "page"},
				State:       AlertStateFiring,
				Value:       "1e+00",
			}},
			Health:         RuleHealthGood,
			LastError:      "Unknown",
			EvaluationTime: 1,
			LastEvaluation: time.Now().Round(time.Millisecond).UTC(),
			State:          "state",
		},
	})
	if err != nil {
		b.Fatal(err)
	}
	b.Log("alerting:", string(alertingRuleJSON))

	recordingRuleJSON, err := json.Marshal(struct {
		Type          RuleType `json:"type"`
		RecordingRule `json:""`
	}{
		Type: RuleTypeRecording,
		RecordingRule: RecordingRule{
			Name:           "job:http_inprogress_requests:sum",
			Query:          "sum(http_inprogress_requests) by (job)",
			Labels:         model.LabelSet{"severity": "page"},
			Health:         RuleHealthGood,
			LastError:      "Unknown",
			EvaluationTime: 1,
			LastEvaluation: time.Now().Round(time.Millisecond).UTC(),
		},
	})
	if err != nil {
		b.Fatal(err)
	}
	b.Log("recording:", string(recordingRuleJSON))

	data := []byte(`{
"name":"myname","file":"myfile","interval":0.0000005,"rules":[` +
		string(alertingRuleJSON) + strings.Repeat(","+string(alertingRuleJSON), 100) + strings.Repeat(","+string(recordingRuleJSON), 100) +
		`]}`)

	b.Run("streaming", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := json.NewDecoder(bytes.NewReader(data)).Decode(&RuleGroup{}); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := json.Unmarshal(data, &RuleGroup{}); err != nil {
				b.Fatal(err)
			}
		}
	})
}
