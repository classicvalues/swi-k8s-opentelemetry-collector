// Copyright 2022 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
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

import (
	"context"
	"testing"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"

	internaldata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus"
)

func TestSwMetricsTransformProcessor(t *testing.T) {
	for _, test := range standardTests {
		t.Run(test.name, func(t *testing.T) {
			next := new(consumertest.MetricsSink)

			p := newSwMetricsTransformProcessor(zap.NewExample(), test.transforms)

			mtp, err := processorhelper.NewMetricsProcessor(
				&Config{
					ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(typeStr)),
				},
				next,
				p.processMetrics,
				processorhelper.WithCapabilities(consumerCapabilities))
			require.NoError(t, err)

			caps := mtp.Capabilities()
			assert.Equal(t, true, caps.MutatesData)
			ctx := context.Background()

			// process
			cErr := mtp.ConsumeMetrics(context.Background(), internaldata.OCToMetrics(nil, nil, test.in))
			assert.NoError(t, cErr)

			// get and check results
			got := next.AllMetrics()
			require.Equal(t, 1, len(got))
			_, _, actualOutMetrics := internaldata.ResourceMetricsToOC(got[0].ResourceMetrics().At(0))
			require.Equal(t, len(test.out), len(actualOutMetrics))

			for idx, out := range test.out {
				actualOut := actualOutMetrics[idx]
				if diff := cmp.Diff(actualOut, out, protocmp.Transform()); diff != "" {
					t.Errorf("Unexpected difference:\n%v", diff)
				}
			}

			assert.NoError(t, mtp.Shutdown(ctx))
		})
	}
}

func BenchmarkSwMetricsTransformProcessorRenameMetrics(b *testing.B) {
	const metricCount = 1000

	transforms := []internalTransform{
		{
			MetricIncludeFilter: internalFilterStrict{include: "metric"},
			Action:              Insert,
			NewName:             "new/metric1",
		},
	}

	in := make([]*metricspb.Metric, metricCount)
	for i := 0; i < metricCount; i++ {
		in[i] = metricBuilder().setName("metric1").build()
	}
	p := newSwMetricsTransformProcessor(nil, transforms)
	mtp, _ := processorhelper.NewMetricsProcessor(&Config{}, consumertest.NewNop(), p.processMetrics)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mtp.ConsumeMetrics(context.Background(), internaldata.OCToMetrics(nil, nil, in))
	}
}
