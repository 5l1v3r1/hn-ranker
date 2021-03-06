package hnclass

type TrainingData struct {
	Vectors []FeatureVector
	Classes []int
}

type Classifier interface {
	SerializerType() string
	Serialize() []byte
	Classify(vec FeatureVector) int
}

type TrainableClassifier interface {
	Classifier
	Train(training, crossValidation *TrainingData)
}

type ClassifierMaker func(m *FeatureMap, classCount int) (TrainableClassifier, error)
type Deserializer func(m *FeatureMap, d []byte) (Classifier, error)

var ClassifierMakers = map[string]ClassifierMaker{
	"neuralnet": func(m *FeatureMap, cc int) (TrainableClassifier, error) {
		return NewNeuralNet(m, cc)
	},
}

var Deserializers = map[string]Deserializer{
	"neuralnet": func(m *FeatureMap, d []byte) (Classifier, error) {
		return DeserializeNeuralNet(m, d)
	},
}
