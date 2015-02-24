package vom

import (
	"bytes"
	"fmt"
	"testing"

	"v.io/v23/vom/testdata"
)

func TestBinaryEncoder(t *testing.T) {
	for _, test := range testdata.Tests {
		name := test.Name + " [vdl.Value]"
		testEncode(t, name, test.Value, test.Hex)

		// Convert into Go value for the rest of our tests.
		goValue, err := toGoValue(test.Value)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}

		name = test.Name + " [go value]"
		testEncode(t, name, goValue, test.Hex)
	}
}

func testEncode(t *testing.T, name string, value interface{}, hex string) {
	var buf bytes.Buffer
	encoder, err := NewEncoder(&buf)
	if err != nil {
		t.Errorf("%s: NewEncoder failed: %v", name, err)
		return
	}
	if err := encoder.Encode(value); err != nil {
		t.Errorf("%s: binary Encode(%#v) failed: %v", name, value, err)
		return
	}
	got, want := fmt.Sprintf("%x", buf.String()), hex
	match, err := matchHexPat(got, want)
	if err != nil {
		t.Error(err)
	}
	if !match {
		t.Errorf("%s: binary Encode(%#v)\nGOT  %s\nWANT %s", name, value, got, want)
	}
}
