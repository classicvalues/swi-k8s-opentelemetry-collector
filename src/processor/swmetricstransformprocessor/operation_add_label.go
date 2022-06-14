// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor
// Changes customizing the original processor:
//	- removal of actions: toggle_scalar_data_type, experimental_scale_value, aggregate_labels, aggregate_label_values
//	- add custom action "filter_datapoints"
//	- rename types and functions to match the processor name

package swmetricstransformprocessor

import metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"

func (mtp *swMetricsTransformProcessor) addLabelOp(metric *metricspb.Metric, op internalOperation) {
	var lb = metricspb.LabelKey{
		Key: op.configOperation.NewLabel,
	}
	metric.MetricDescriptor.LabelKeys = append(metric.MetricDescriptor.LabelKeys, &lb)
	for _, ts := range metric.Timeseries {
		lv := &metricspb.LabelValue{
			Value:    op.configOperation.NewValue,
			HasValue: true,
		}
		ts.LabelValues = append(ts.LabelValues, lv)
	}
}
