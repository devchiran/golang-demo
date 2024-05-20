package mock

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer implements the Tracer interface for mocking purposes.
type Tracer struct {
	StartFn func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}

// Start implements the Start method for mocking purposes.
func (t Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.StartFn(ctx, spanName, opts...)
}

// Span implements the Span interface for mocking purposes.
type Span struct {
	EndFn            func(options ...trace.SpanEndOption)
	AddEventFn       func(name string, options ...trace.EventOption)
	IsRecordingFn    func() bool
	RecordErrorFn    func(err error, options ...trace.EventOption)
	SpanContextFn    func() trace.SpanContext
	SetStatusFn      func(code codes.Code, description string)
	SetNameFn        func(name string)
	SetAttributesFn  func(kv ...attribute.KeyValue)
	TracerProviderFn func() trace.TracerProvider
}

// End implements the End method for mocking purposes.
func (s Span) End(options ...trace.SpanEndOption) {
	s.EndFn(options...)
}

// AddEvent implements the AddEvent method for mocking purposes.
func (s Span) AddEvent(name string, options ...trace.EventOption) {
	s.AddEventFn(name, options...)
}

// IsRecording implements the IsRecording method for mocking purposes.
func (s Span) IsRecording() bool {
	return s.IsRecordingFn()
}

// RecordError implements the RecordError method for mocking purposes.
func (s Span) RecordError(err error, options ...trace.EventOption) {
	s.RecordErrorFn(err, options...)
}

// SpanContext implements the SpanContext method for mocking purposes.
func (s Span) SpanContext() trace.SpanContext {
	return s.SpanContextFn()
}

// SetStatus implements the SetStatus method for mocking purposes.
func (s Span) SetStatus(code codes.Code, description string) {
	s.SetStatusFn(code, description)
}

// SetName implements the SetName method for mocking purposes.
func (s Span) SetName(name string) {
	s.SetNameFn(name)
}

// SetAttributes implements the SetAttributes method for mocking purposes.
func (s Span) SetAttributes(kv ...attribute.KeyValue) {
	s.SetAttributesFn(kv...)
}

// TracerProvider implements the TracerProvider method for mocking purposes.
func (s Span) TracerProvider() trace.TracerProvider {
	return s.TracerProviderFn()
}
