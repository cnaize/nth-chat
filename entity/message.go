package entity

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Message struct {
	From string
	To   string
	Text string
}

func (m *Message) Marshal() ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buff).Encode(m); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	
	return buff.Bytes(), nil
}

func (m *Message) Unmarshal(data []byte) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(m)
}
