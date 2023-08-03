package main

import (
	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type RatioBasedSampler struct {
	innerSampler        tracesdk.Sampler
	sampleRateAttribute attribute.KeyValue
}

func NewRatioBasedSampler(fraction float64) RatioBasedSampler {
	innerSampler := tracesdk.TraceIDRatioBased(fraction)
	return RatioBasedSampler{
		innerSampler:        innerSampler,
		sampleRateAttribute: attribute.Float64("X-SampleRatio", fraction),
	}
}

func (ds RatioBasedSampler) ShouldSample(parameters tracesdk.SamplingParameters) tracesdk.SamplingResult {
	sampler := ds.innerSampler
	result := sampler.ShouldSample(parameters)
	if result.Decision == tracesdk.RecordAndSample {
		result.Attributes = append(result.Attributes, ds.sampleRateAttribute)
	}
	return result
}

func (ds RatioBasedSampler) Description() string {
	return "Ratio Based Sampler which gives information about sampling ratio"
}
