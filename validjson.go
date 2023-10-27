package validjson

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mvndaai/ctxerr"
)

const (
	ErrorCodeMissingBody   = "MISSING_BODY"
	ErrorCodeMalformedBody = "MALFORMED_BODY"
	ErrorCodeInvalidBody   = "INVALID_BODY"
)

const (
	fieldKeyBody = "body"
)

type Validator interface{ Validate() error }
type ContextValidator interface{ Validate(context.Context) error }
type Normalizer interface{ Normalize() }
type ContextNormalizer interface{ Normalize(context.Context) }
type Redactor interface{ Redacted() any }
type ContextRedactor interface{ Redacted(context.Context) any }

func Unmarshal(ctx context.Context, b []byte, i any) error {
	if len(b) == 0 {
		return ctxerr.NewHTTP(ctx, ErrorCodeMissingBody, "Missing request body", http.StatusBadRequest, "missing request body")
	}

	err := json.Unmarshal(b, i)
	if err != nil {
		ctx = ctxerr.SetField(ctx, fieldKeyBody, string(b))
		return ctxerr.WrapHTTP(ctx, err, ErrorCodeMalformedBody, "Malformed request body", http.StatusBadRequest, "bad request body")
	}

	switch v := i.(type) {
	case ContextNormalizer:
		v.Normalize(ctx)
	case Normalizer:
		v.Normalize()
	}

	switch v := i.(type) {
	case ContextValidator:
		err = v.Validate(ctx)
	case Validator:
		err = v.Validate()
	}
	if err != nil {
		ctx = ctxerr.SetField(ctx, fieldKeyBody, TryToRedact(ctx, i))
		return ctxerr.WrapHTTP(ctx, err, ErrorCodeInvalidBody, err.Error(), http.StatusBadRequest)
	}

	return nil
}

func TryToRedact(ctx context.Context, a any) any {
	switch v := a.(type) {
	case Redactor:
		return v.Redacted()
	case ContextRedactor:
		return v.Redacted(ctx)
	}
	return a
}
