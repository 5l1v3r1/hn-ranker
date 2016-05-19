package hnclass

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
)

var errBufferUnderflow = errors.New("buffer underflow")

var serializeByteOrder = binary.LittleEndian

func Serialize(c Classifier, m *FeatureMap) []byte {
	featureData, _ := json.Marshal(m)

	var b bytes.Buffer
	binary.Write(&b, serializeByteOrder, uint64(len(featureData)))
	b.Write(featureData)

	t := []byte(c.SerializerType())
	binary.Write(&b, serializeByteOrder, uint64(len(t)))
	b.Write(t)
	b.Write(c.Serialize())

	return b.Bytes()
}

func Deserialize(d []byte) (Classifier, *FeatureMap, error) {
	b := bytes.NewBuffer(d)

	var lenField uint64
	if err := binary.Read(b, serializeByteOrder, &lenField); err != nil {
		return nil, nil, err
	}

	featureData := make([]byte, int(lenField))
	if n, _ := b.Read(featureData); n < len(featureData) {
		return nil, nil, errBufferUnderflow
	}

	var features FeatureMap
	if err := json.Unmarshal(featureData, &features); err != nil {
		return nil, nil, err
	}

	if err := binary.Read(b, serializeByteOrder, &lenField); err != nil {
		return nil, nil, err
	}
	nameData := make([]byte, int(lenField))
	if n, _ := b.Read(nameData); n < len(nameData) {
		return nil, nil, errBufferUnderflow
	}

	deserializer, ok := Deserializers[string(nameData)]
	if !ok {
		return nil, nil, fmt.Errorf("unknown classifier type: %s", string(nameData))
	}

	class, err := deserializer(&features, b.Bytes())
	if err != nil {
		return nil, nil, err
	}

	return class, &features, nil
}
