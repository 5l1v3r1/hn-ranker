package main

import (
	"encoding/json"

	"github.com/unixpickle/weakai/neuralnet"
)

type Classifier struct {
	Features *Features
	Network  *neuralnet.Network
}

func DecodeClassifier(d []byte) (*Classifier, error) {
	var s struct {
		Features *Features
		NetData  []byte
	}
	if err := json.Unmarshal(d, &s); err != nil {
		return nil, err
	}
	network, err := neuralnet.DeserializeNetwork(s.NetData)
	if err != nil {
		return nil, err
	}
	return &Classifier{
		Features: s.Features,
		Network:  network,
	}, nil
}

func (c *Classifier) Encode() []byte {
	var s struct {
		Features *Features
		NetData  []byte
	}
	s.Features = c.Features
	s.NetData = c.Network.Serialize()
	data, _ := json.Marshal(&s)
	return data
}
