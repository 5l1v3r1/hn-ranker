package main

type Classifier interface {
	Serialize() []byte
	Classify(vec FeatureVector) int
}

type TrainableClassifier interface {
	Classifier
	Train(vecs []FeatureVector, classes []int)
}

type ClassifierMaker func(m *FeatureMap) (TrainableClassifier, error)
type Deserializer func(m *FeatureMap, d []byte) (Classifier, error)

var ClassifierMakers = map[string]ClassifierMaker{
	"neuralnet": func(m *FeatureMap) (TrainableClassifier, error) {
		return NewNeuralNet(m)
	},
}

var Deserializers = map[string]Deserializer{
	"neuralnet": func(m *FeatureMap, d []byte) (Classifier, error) {
		return DeserializeNeuralNet(m, d)
	},
}
