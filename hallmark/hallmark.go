// Package hallmark emits OpenTelemetry hallmark spans that carry a
// judge's observations about one artefact.
//
// Schema reference: werkhaus/docs/plans/2026-04-23-hallmark-span-schema.md.
// A hallmark span records evidence, not verdicts: downstream aggregators
// derive ABI (Ability/Benevolence/Integrity) from streams of hallmarks
// joined to the producing worker's identity via artefact -> commission.
package hallmark

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	spanName    = "hallmark"
	tracerName  = "github.com/benaskins/axon-hand/hallmark"
	idPrefix    = "hlm_"
	obsPrefix   = "observation."
	suffixDet   = "-det"
	suffixLLM   = "-llm"
)

// Phase enumerates the build stages a hallmark may target.
type Phase string

const (
	PhasePlan Phase = "plan"
	PhaseCode Phase = "code"
	PhaseScan Phase = "scan"
	PhaseSeal Phase = "seal"
	PhaseShip Phase = "ship"
)

var validPhases = map[Phase]bool{
	PhasePlan: true, PhaseCode: true, PhaseScan: true,
	PhaseSeal: true, PhaseShip: true,
}

// Source identifies the evaluator channel producing a hallmark.
// Format <craft>-<mode> where mode is llm or det.
type Source string

const (
	SourceCodeLeadLLM Source = "code-lead-llm"
	SourcePlanLeadLLM Source = "plan-lead-llm"
	SourceScanLeadDet Source = "scan-lead-det"
	SourceTestRunner  Source = "test-runner"
	SourceHumanReview Source = "human-review"
)

var validSources = map[Source]bool{
	SourceCodeLeadLLM: true,
	SourcePlanLeadLLM: true,
	SourceScanLeadDet: true,
	SourceTestRunner:  true,
	SourceHumanReview: true,
}

// Judge identifies the evaluator.
type Judge struct {
	Craft   string // e.g. "scan-lead"
	Version string // semver or git sha
	Model   string // LLM name; empty for -det sources
}

// Observation is one signal on the hallmark. Value must be bool, int64,
// float64, or string. Confidence is in [0.0, 1.0].
type Observation struct {
	Name       string
	Value      any
	Confidence float64
	Rationale  string
}

// Input is the complete shape of one hallmark emission.
type Input struct {
	Phase        Phase
	Source       Source
	ArtefactID   string
	Judge        Judge
	BuildID      string // optional
	CommissionID string // optional
	Observations []Observation
}

// Emit validates the input, starts an INTERNAL span named "hallmark" on
// the tracer bound to ctx, populates all schema attributes, and ends
// the span. It returns the generated hallmark id.
func Emit(ctx context.Context, in Input) (string, error) {
	if err := validate(in); err != nil {
		return "", err
	}

	id := newID()
	tracer := otel.Tracer(tracerName)
	_, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	attrs := []attribute.KeyValue{
		attribute.String("hallmark.id", id),
		attribute.String("hallmark.phase", string(in.Phase)),
		attribute.String("hallmark.source", string(in.Source)),
		attribute.String("hallmark.artefact.id", in.ArtefactID),
		attribute.String("hallmark.judge.craft", in.Judge.Craft),
		attribute.String("hallmark.judge.version", in.Judge.Version),
		attribute.String("hallmark.judge.model", in.Judge.Model),
	}
	if in.BuildID != "" {
		attrs = append(attrs, attribute.String("hallmark.build.id", in.BuildID))
	}
	if in.CommissionID != "" {
		attrs = append(attrs, attribute.String("hallmark.commission.id", in.CommissionID))
	}

	for _, o := range in.Observations {
		vAttr, err := observationValueAttr(o.Name+".value", o.Value)
		if err != nil {
			return "", err
		}
		attrs = append(attrs,
			vAttr,
			attribute.Float64(o.Name+".confidence", o.Confidence),
			attribute.String(o.Name+".rationale", o.Rationale),
		)
	}

	span.SetAttributes(attrs...)
	return id, nil
}

func validate(in Input) error {
	if !validPhases[in.Phase] {
		return fmt.Errorf("hallmark: invalid phase %q", in.Phase)
	}
	if !validSources[in.Source] {
		return fmt.Errorf("hallmark: invalid source %q", in.Source)
	}
	if in.ArtefactID == "" {
		return errors.New("hallmark: artefact id required")
	}
	if in.Judge.Craft == "" {
		return errors.New("hallmark: judge.craft required")
	}
	if in.Judge.Version == "" {
		return errors.New("hallmark: judge.version required")
	}
	if err := validateModelForSource(in.Source, in.Judge.Model); err != nil {
		return err
	}
	if len(in.Observations) == 0 {
		return errors.New("hallmark: observations must not be empty")
	}
	for i, o := range in.Observations {
		if !strings.HasPrefix(o.Name, obsPrefix) {
			return fmt.Errorf("hallmark: observation.name[%d] %q must begin with %q", i, o.Name, obsPrefix)
		}
		if o.Confidence < 0.0 || o.Confidence > 1.0 {
			return fmt.Errorf("hallmark: observation[%d] confidence %v not in [0.0, 1.0]", i, o.Confidence)
		}
		if !validValueType(o.Value) {
			return fmt.Errorf("hallmark: observation.value[%d] type %T unsupported", i, o.Value)
		}
	}
	return nil
}

func validateModelForSource(src Source, model string) error {
	s := string(src)
	switch {
	case strings.HasSuffix(s, suffixDet):
		if model != "" {
			return fmt.Errorf("hallmark: judge.model must be empty for -det source %q", src)
		}
	case strings.HasSuffix(s, suffixLLM):
		if model == "" {
			return fmt.Errorf("hallmark: judge.model required for -llm source %q", src)
		}
	}
	return nil
}

func validValueType(v any) bool {
	switch v.(type) {
	case bool, int64, float64, string:
		return true
	default:
		return false
	}
}

func observationValueAttr(key string, v any) (attribute.KeyValue, error) {
	switch x := v.(type) {
	case bool:
		return attribute.Bool(key, x), nil
	case int64:
		return attribute.Int64(key, x), nil
	case float64:
		return attribute.Float64(key, x), nil
	case string:
		return attribute.String(key, x), nil
	default:
		return attribute.KeyValue{}, fmt.Errorf("hallmark: unsupported observation value type %T", v)
	}
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return idPrefix + "000000000000000000000000000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40 // v4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%s%08x-%04x-%04x-%04x-%012x",
		idPrefix, b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
