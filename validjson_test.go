package validjson_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/mvndaai/ctxerr"
	"github.com/mvndaai/validjson"
	"github.com/stretchr/testify/assert"
)

const (
	normalizedValue = "normalized"
	redactedValue   = "redacted"
)

type normal struct{ V string }

func (s normal) Validate() error { return nil }
func (s *normal) Normalize()     { s.V = normalizedValue }

type normalContext struct{ V string }

func (s normalContext) Validate(_ context.Context) error { return nil }
func (s *normalContext) Normalize(context.Context)       { s.V = normalizedValue }

type redact struct{ V string }

func (s redact) Validate() error { return ctxerr.New(context.Background(), "", "redact validate") }
func (s *redact) Redact() any    { s.V = redactedValue; return s }

type redactContext struct{ V string }

func (s redactContext) Validate(ctx context.Context) error {
	return ctxerr.New(ctx, "", "redact context validate")
}
func (s *redactContext) Redact(context.Context) any { s.V = redactedValue; return s }

type noTransforms struct{ V string }

type noRedact struct{ V string }

func (s noRedact) Validate(ctx context.Context) error {
	return ctxerr.New(ctx, "", "no redact validate")
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name          string
		body          []byte
		pointer       any
		normalized    any
		errorContains string
		redacted      any
	}{
		{
			name:          "normal",
			body:          []byte(`{"v": "a"}`),
			pointer:       &normal{},
			normalized:    &normal{V: normalizedValue},
			errorContains: "",
			redacted:      nil,
		},
		{
			name:          "normal context",
			body:          []byte(`{"v": "a"}`),
			pointer:       &normalContext{},
			normalized:    &normalContext{V: normalizedValue},
			errorContains: "",
			redacted:      nil,
		},
		{
			name:          "redacted",
			body:          []byte(`{"v": "a"}`),
			pointer:       &redact{},
			normalized:    nil,
			errorContains: "failed validation : redact validate",
			redacted:      &redact{V: redactedValue},
		},
		{
			name:          "redacted context",
			body:          []byte(`{"v": "a"}`),
			pointer:       &redactContext{},
			normalized:    nil,
			errorContains: "failed validation : redact context validate",
			redacted:      &redactContext{V: redactedValue},
		},
		{
			name:          "no transforms",
			body:          []byte(`{"v": "a"}`),
			pointer:       &noTransforms{},
			normalized:    &noTransforms{V: "a"},
			errorContains: "",
			redacted:      &redactContext{V: redactedValue},
		},
		{
			name:          "no redact",
			body:          []byte(`{"v": "a"}`),
			pointer:       &noRedact{},
			normalized:    nil,
			errorContains: "failed validation : no redact validate",
			redacted:      &redactContext{V: "a"},
		},
		{
			name:          "missing body",
			body:          nil,
			pointer:       &noTransforms{},
			normalized:    nil,
			errorContains: "missing request body",
			redacted:      nil,
		},
		{
			name:          "malformed body",
			body:          []byte(`{"v": 1}`),
			pointer:       &redact{},
			normalized:    nil,
			errorContains: "bad request body",
			redacted:      `{"v": 1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loop := map[string]func() error{
				"byte": func() error { return validjson.Unmarshal(context.Background(), tt.body, tt.pointer) },
				"read_closer": func() error {
					r := io.NopCloser(strings.NewReader(string(tt.body)))
					return validjson.UnmarshalReadCloser(context.Background(), r, tt.pointer)
				},
			}

			for name, f := range loop {
				t.Run(name, func(t *testing.T) {
					err := f()
					if err != nil {
						es := err.Error()
						_ = es
						assert.Contains(t, err.Error(), tt.errorContains)
						f := ctxerr.AllFields(err)
						body := f["body"]
						assert.EqualValues(t, tt.redacted, body)
						return
					}
					assert.Equal(t, err == nil, tt.errorContains == "")
					assert.EqualValues(t, tt.pointer, tt.normalized)
				})
			}

		})
	}
}
