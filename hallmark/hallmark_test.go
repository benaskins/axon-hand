package hallmark

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func installRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
		otel.SetTracerProvider(prev)
	})
	return rec
}

func attrMap(attrs []attribute.KeyValue) map[string]attribute.Value {
	m := make(map[string]attribute.Value, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value
	}
	return m
}

func validInput() Input {
	return Input{
		Phase:      PhaseScan,
		Source:     SourceScanLeadDet,
		ArtefactID: "art_1234",
		Judge: Judge{
			Craft:   "scan-lead",
			Version: "0.1.0",
			Model:   "",
		},
		Observations: []Observation{
			{
				Name:       "observation.tests_passed",
				Value:      true,
				Confidence: 1.0,
				Rationale:  "3/3 unit green",
			},
		},
	}
}

func TestEmit_CreatesHallmarkSpan(t *testing.T) {
	rec := installRecorder(t)

	id, err := Emit(context.Background(), validInput())
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}
	if !strings.HasPrefix(id, "hlm_") {
		t.Errorf("id = %q, want prefix hlm_", id)
	}

	spans := rec.Ended()
	if len(spans) != 1 {
		t.Fatalf("spans=%d, want 1", len(spans))
	}
	s := spans[0]
	if s.Name() != "hallmark" {
		t.Errorf("span name = %q, want hallmark", s.Name())
	}
	if s.SpanKind() != oteltrace.SpanKindInternal {
		t.Errorf("span kind = %v, want Internal", s.SpanKind())
	}
}

func TestEmit_SetsEnvelopeAttributes(t *testing.T) {
	rec := installRecorder(t)

	in := validInput()
	in.BuildID = "bld_abc"
	in.CommissionID = "cmn_def"

	id, err := Emit(context.Background(), in)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}

	attrs := attrMap(rec.Ended()[0].Attributes())

	cases := map[string]string{
		"hallmark.id":            id,
		"hallmark.phase":         "scan",
		"hallmark.source":        "scan-lead-det",
		"hallmark.artefact.id":   "art_1234",
		"hallmark.build.id":      "bld_abc",
		"hallmark.commission.id": "cmn_def",
		"hallmark.judge.craft":   "scan-lead",
		"hallmark.judge.version": "0.1.0",
		"hallmark.judge.model":   "",
	}
	for k, want := range cases {
		v, ok := attrs[k]
		if !ok {
			t.Errorf("missing attr %q", k)
			continue
		}
		if v.AsString() != want {
			t.Errorf("attr %q = %q, want %q", k, v.AsString(), want)
		}
	}
}

func TestEmit_OmitsOptionalAttributesWhenUnset(t *testing.T) {
	rec := installRecorder(t)

	_, err := Emit(context.Background(), validInput())
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}

	attrs := attrMap(rec.Ended()[0].Attributes())
	for _, k := range []string{"hallmark.build.id", "hallmark.commission.id"} {
		if _, present := attrs[k]; present {
			t.Errorf("attr %q should be absent when unset", k)
		}
	}
}

func TestEmit_ObservationAttributes_Bool(t *testing.T) {
	rec := installRecorder(t)

	_, err := Emit(context.Background(), validInput())
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}

	attrs := attrMap(rec.Ended()[0].Attributes())
	if v, ok := attrs["observation.tests_passed.value"]; !ok {
		t.Fatalf("missing observation value attr")
	} else if v.Type() != attribute.BOOL || v.AsBool() != true {
		t.Errorf("value = %v, want bool true", v.Emit())
	}
	if v, ok := attrs["observation.tests_passed.confidence"]; !ok {
		t.Fatalf("missing confidence attr")
	} else if v.Type() != attribute.FLOAT64 || v.AsFloat64() != 1.0 {
		t.Errorf("confidence = %v, want 1.0", v.Emit())
	}
	if v, ok := attrs["observation.tests_passed.rationale"]; !ok {
		t.Fatalf("missing rationale attr")
	} else if v.Type() != attribute.STRING || v.AsString() != "3/3 unit green" {
		t.Errorf("rationale = %v, want 3/3 unit green", v.Emit())
	}
}

func TestEmit_ObservationAttributes_AllTypes(t *testing.T) {
	rec := installRecorder(t)

	in := validInput()
	in.Observations = []Observation{
		{Name: "observation.slop_em_dash_count", Value: int64(5), Confidence: 1.0, Rationale: "5 em-dashes in README.md"},
		{Name: "observation.spec_match", Value: float64(0.8), Confidence: 0.6, Rationale: "fields match, auth pattern deviates"},
		{Name: "observation.consumer_disposition", Value: "accept", Confidence: 0.9, Rationale: "output meets brief"},
	}

	_, err := Emit(context.Background(), in)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}
	attrs := attrMap(rec.Ended()[0].Attributes())

	if v := attrs["observation.slop_em_dash_count.value"]; v.Type() != attribute.INT64 || v.AsInt64() != 5 {
		t.Errorf("int64 value = %v", v.Emit())
	}
	if v := attrs["observation.spec_match.value"]; v.Type() != attribute.FLOAT64 || v.AsFloat64() != 0.8 {
		t.Errorf("float64 value = %v", v.Emit())
	}
	if v := attrs["observation.consumer_disposition.value"]; v.Type() != attribute.STRING || v.AsString() != "accept" {
		t.Errorf("string value = %v", v.Emit())
	}
}

func TestEmit_Validates(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*Input)
		want string
	}{
		{"missing phase", func(in *Input) { in.Phase = "" }, "phase"},
		{"invalid phase", func(in *Input) { in.Phase = "bogus" }, "phase"},
		{"missing source", func(in *Input) { in.Source = "" }, "source"},
		{"invalid source", func(in *Input) { in.Source = "bogus-llm" }, "source"},
		{"missing artefact id", func(in *Input) { in.ArtefactID = "" }, "artefact"},
		{"missing judge craft", func(in *Input) { in.Judge.Craft = "" }, "judge.craft"},
		{"missing judge version", func(in *Input) { in.Judge.Version = "" }, "judge.version"},
		{"det source with model", func(in *Input) { in.Judge.Model = "some-model" }, "judge.model"},
		{"no observations", func(in *Input) { in.Observations = nil }, "observations"},
		{"observation name bad prefix", func(in *Input) {
			in.Observations = []Observation{{Name: "foo.bar", Value: true, Confidence: 1.0}}
		}, "observation.name"},
		{"observation confidence out of range", func(in *Input) {
			in.Observations = []Observation{{Name: "observation.x", Value: true, Confidence: 1.5}}
		}, "confidence"},
		{"observation unsupported value type", func(in *Input) {
			in.Observations = []Observation{{Name: "observation.x", Value: []int{1}, Confidence: 1.0}}
		}, "observation.value"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			installRecorder(t)
			in := validInput()
			tc.mut(&in)
			_, err := Emit(context.Background(), in)
			if err == nil {
				t.Fatalf("Emit: want error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestEmit_LLMSourceRequiresModel(t *testing.T) {
	installRecorder(t)
	in := validInput()
	in.Source = SourceCodeLeadLLM
	in.Judge.Craft = "code-lead"
	in.Judge.Model = "" // missing

	_, err := Emit(context.Background(), in)
	if err == nil || !strings.Contains(err.Error(), "judge.model") {
		t.Fatalf("want judge.model error, got %v", err)
	}
}

func TestEmit_LLMSourceAcceptsModel(t *testing.T) {
	installRecorder(t)
	in := validInput()
	in.Source = SourceCodeLeadLLM
	in.Judge.Craft = "code-lead"
	in.Judge.Model = "claude-sonnet-4-6"

	_, err := Emit(context.Background(), in)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}
}
