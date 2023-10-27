package validjson_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

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

func TestValid(t *testing.T) {
	body := []byte(`{
		"upper": "up",
		"trimmed": " remove spaces ",
		"secret": "caution",
		"required": true
	}`)

	var ex example
	err := validjson.Unmarshal(context.Background(), body, &ex)
	require.NoError(t, err)

	t.Logf("%#v", ex)
	//t.Error("show the log")
}

type Fielder interface {
	Fields() map[string]any
}

func TestInvalid(t *testing.T) {
	body := []byte(`{
		"upper": "up",
		"trimmed": " remove spaces ",
		"secret": "caution",
		"required": false
	}`)

	var ex example
	err := validjson.Unmarshal(context.Background(), body, &ex)
	require.Error(t, err)

	fields := err.(Fielder).Fields()
	b, _ := json.MarshalIndent(fields, "", "\t")
	t.Log("err:", err.Error())
	t.Log("fields:", string(b))
	//t.Error("show the log")
}
