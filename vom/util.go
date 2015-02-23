package vom

import "bytes"

// Encode encodes the provided value using a new instance of a VOM encoder.
func Encode(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder, err := NewEncoder(&buf)
	if err != nil {
		return nil, err
	}
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode VOM-decodes the given data into the provided value using a new
// instance of a VOM decoder.
func Decode(data []byte, valptr interface{}) error {
	decoder, err := NewDecoder(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return decoder.Decode(valptr)
}
