package bottle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseArgsForLabelKeyValue_Labels(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"valid", "key=value", ""},
		{"invalid", "key= value", "a valid label must be an empty string or consist of alphanumeric characters"},
		{"missing equals", "key value", "invalid format"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			key, value, err := parseLabel(c.input)
			if c.wantErr != "" {
				assert.ErrorContains(t, err, c.wantErr)
				return
			}
			assert.Equal(t, "key", key)
			assert.Equal(t, "value", value)
		})
	}
}
