package bottle

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseArgsForAnnotKeyValue(t *testing.T) {

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{"normal Key value", "Key", "value", false},
		{"Key with space", "image ", "redis", true},
		{"base64 value", "data", "PWhlbGxvd29ybGQ=", false},
		{"value with number only", "size", "1024", false},
		{"value is JSON", "json", `{"name":"John", "age":30, "car":null}`, false},
		{"value with space after =", "Key", " value", false},
		{"bad Key", "Key+()", " value", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFmt := fmt.Sprintf("%s=%s", tt.key, tt.value)
			key, value, err := parseAnnotation(inputFmt)

			if tt.wantErr {
				assert.ErrorContains(t, err, "name part must consist of alphanumeric characters")
				return
			}
			assert.Equalf(t, tt.key, key, "Expected Key: %s | Found Key: %s", tt.key, key)
			assert.Equalf(t, tt.value, value, "Expected value: %s | Found value: %s", tt.key, key)
		})
	}
}
