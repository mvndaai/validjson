package validjson_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/mvndaai/ctxerr"
	"github.com/mvndaai/validjson"
	"github.com/stretchr/testify/require"
)

type example struct {
	Upper    string `json:"upper"`
	Trimmed  string `json:"trimmed"`
	Secret   string `json:"secret"`
	Required bool   `json:"required"`
}

func (ex *example) Normalize() {
	ex.Upper = strings.ToUpper(ex.Upper)
	ex.Trimmed = strings.TrimSpace(ex.Trimmed)
}

func (ex example) Validate() error {
	if !ex.Required {
		return errors.New("'required' attr needs to be true")
	}
	return nil
}

func (ex example) Redacted(_ context.Context) any {
	ex.Secret = "*****"
	return ex
}

type Fielder interface {
	Fields() map[string]any
}

func TestExamples(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		expectError bool
	}{
		{
			name: "valid",
			body: `{
				"upper": "up",
				"trimmed": " remove spaces ",
				"secret": "caution",
				"required": true
			}`,
			expectError: false,
		},
		{
			name:        "empty",
			body:        ``,
			expectError: true,
		},
		{
			name: "invalid",
			body: `{
				"upper": "up",
				"trimmed": " remove spaces ",
				"secret": "caution",
				"required": false
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			ctx = ctxerr.SetField(ctx, "pathParam", "id")

			var ex example
			err := validjson.Unmarshal(ctx, []byte(tt.body), &ex)
			require.Equal(t, err != nil, tt.expectError)

			t.Logf("body %#v", ex)
			if err != nil {
				fields := err.(Fielder).Fields()
				b, _ := json.MarshalIndent(fields, "", "\t")
				t.Log("err:", err.Error())
				t.Log("fields:", string(b))
			}

		})
	}
}
