package validjson_test

import (
	"context"
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

func (s redact) Validate() error { return ctxerr.New(context.Background(), "") }
func (s *redact) Redact() any    { s.V = redactedValue; return s }

type redactContext struct{ V string }

func (s redactContext) Validate(ctx context.Context) error { return ctxerr.New(ctx, "") }
func (s *redactContext) Redact(context.Context) any        { s.V = redactedValue; return s }

type noTransforms struct{ V string }

type noRedact struct{ V string }

func (s noRedact) Validate(ctx context.Context) error { return ctxerr.New(ctx, "") }

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name       string
		body       []byte
		pointer    any
		normalized any
		errorCode  string
		redacted   any
	}{
		{
			name:       "normal",
			body:       []byte(`{"v": "a"}`),
			pointer:    &normal{},
			normalized: &normal{V: normalizedValue},
			errorCode:  "",
			redacted:   nil,
		},
		{
			name:       "normal context",
			body:       []byte(`{"v": "a"}`),
			pointer:    &normalContext{},
			normalized: &normalContext{V: normalizedValue},
			errorCode:  "",
			redacted:   nil,
		},
		{
			name:       "redacted",
			body:       []byte(`{"v": "a"}`),
			pointer:    &redact{},
			normalized: nil,
			errorCode:  validjson.ErrorCodeInvalidBody,
			redacted:   &redact{V: redactedValue},
		},
		{
			name:       "redacted context",
			body:       []byte(`{"v": "a"}`),
			pointer:    &redactContext{},
			normalized: nil,
			errorCode:  validjson.ErrorCodeInvalidBody,
			redacted:   &redactContext{V: redactedValue},
		},
		{
			name:       "no transforms",
			body:       []byte(`{"v": "a"}`),
			pointer:    &noTransforms{},
			normalized: &noTransforms{V: "a"},
			errorCode:  "",
			redacted:   &redactContext{V: redactedValue},
		},
		{
			name:       "no redact",
			body:       []byte(`{"v": "a"}`),
			pointer:    &noRedact{},
			normalized: nil,
			errorCode:  validjson.ErrorCodeInvalidBody,
			redacted:   &redactContext{V: "a"},
		},
		{
			name:       "missing body",
			body:       nil,
			pointer:    &noTransforms{},
			normalized: nil,
			errorCode:  validjson.ErrorCodeMissingBody,
			redacted:   nil,
		},
		{
			name:       "malformed body",
			body:       []byte(`{"v": 1}`),
			pointer:    &redact{},
			normalized: nil,
			errorCode:  validjson.ErrorCodeMalformedBody,
			redacted:   `{"v": 1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validjson.Unmarshal(context.Background(), tt.body, tt.pointer)
			if err != nil {
				f := ctxerr.AllFields(err)
				errorCode := f[ctxerr.FieldKeyCode]
				assert.Equal(t, tt.errorCode, errorCode)

				body := f["body"]
				assert.EqualValues(t, tt.redacted, body)
				return
			}

			assert.EqualValues(t, tt.pointer, tt.normalized)
		})
	}
}
