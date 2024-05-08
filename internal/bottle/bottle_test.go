package bottle

import (
	"testing"

	cfgdef "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"

	"github.com/stretchr/testify/assert"
)

// The following mimic a schema definition.
type MockStruct struct {
	StringThing string
	StringSlice []string
	BoolThing   bool
	IntThing    int
	StructSlice []MockSubStruct
}
type MockSubStruct struct {
	SubString      string
	SubStringSlice []string
	SubBool        bool
	SubInt         int
	SubStructSlice []SubSubStruct
}

type SubSubStruct struct {
	SubSubString      string
	SubSubStringSlice []string
	SubSubBool        bool
	SubSubInt         int
}

// MockBottle returns a mock *Bottle struct with a Mock latest.BottleDefinition
func MockBottle() (btl *Bottle) {
	mockBottle := Bottle{
		Definition: cfgdef.NewBottle(),
	}
	return &mockBottle
}

func Test_MockBottle(t *testing.T) {
	mbottle := MockBottle()
	assert.NotNil(t, mbottle)
}
