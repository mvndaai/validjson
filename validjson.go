package validjson

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/mvndaai/ctxerr"
)

const (
	fieldKeyBody = "body"
)

type Validator interface {
	Validate() error
}
type ContextValidator interface {
	Validate(context.Context) error
}
type Normalizer interface {
	Normalize()
}
type ContextNormalizer interface {
	Normalize(context.Context)
}
type Redactor interface {
	Redact() any
}
type ContextRedactor interface {
	Redact(context.Context) any
}

func Unmarshal(ctx context.Context, b []byte, i any) error {
	if len(b) == 0 {
		return ctxerr.NewHTTP(ctx, "", "Missing request body", http.StatusBadRequest, "missing request body")
	}

	err := json.Unmarshal(b, i)
	if err != nil {
		ctx = ctxerr.SetField(ctx, fieldKeyBody, string(b))
		return ctxerr.WrapHTTP(ctx, err, "", "Malformed request body", http.StatusBadRequest, "bad request body")
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
		return ctxerr.WrapHTTP(ctx, err, "", err.Error(), http.StatusBadRequest, "failed validation")
	}

	return nil
}

func UnmarshalReadCloser(ctx context.Context, r io.ReadCloser, i any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return ctxerr.WrapHTTP(ctx, err, "", "Malformed request body", http.StatusBadRequest, "bad request body")
	}
	return Unmarshal(ctx, b, i)
}

func TryToRedact(ctx context.Context, a any) any {
	switch v := a.(type) {
	case Redactor:
		return v.Redact()
	case ContextRedactor:
		return v.Redact(ctx)
	}
	return a
}
