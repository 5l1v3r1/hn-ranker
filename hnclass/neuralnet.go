package hnclass

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/unixpickle/weakai/neuralnet"
)

type NeuralNet struct {
	trainConfig *neuralNetConfig

	featureMap *FeatureMap
	network    *neuralnet.Network
}

func NewNeuralNet(m *FeatureMap, classCount int) (*NeuralNet, error) {
	config, err := getNeuralNetConfig()
	if err != nil {
		return nil, err
	}

	network, err := neuralnet.NewNetwork([]neuralnet.LayerPrototype{
		&neuralnet.DenseParams{
			Activation:  neuralnet.Sigmoid{},
			InputCount:  m.VectorSize(),
			OutputCount: config.HiddenCount,
		},
		&neuralnet.DenseParams{
			Activation:  neuralnet.Sigmoid{},
			InputCount:  config.HiddenCount,
			OutputCount: classCount,
		},
	})
	if err != nil {
		return nil, err
	}
	network.SetInput(make([]float64, m.VectorSize()))
	network.SetDownstreamGradient(make([]float64, len(network.Output())))

	return &NeuralNet{
		trainConfig: config,
		featureMap:  m,
		network:     network,
	}, nil
}

func DeserializeNeuralNet(m *FeatureMap, d []byte) (*NeuralNet, error) {
	net, err := neuralnet.DeserializeNetwork(d)
	if err != nil {
		return nil, err
	}
	return &NeuralNet{featureMap: m, network: net}, nil
}

func (n *NeuralNet) Train(vecs []FeatureVector, classes []int) {
	killChan := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		signal.Stop(c)
		fmt.Println("\nCaught interrupt. Ctrl+C again to terminate.")
		close(killChan)
	}()
	log.Println("Press Ctrl+C to finish training.")
	n.train(vecs, classes, killChan)
}

func (n *NeuralNet) Serialize() []byte {
	return n.network.Serialize()
}

func (n *NeuralNet) SerializerType() string {
	return "neuralnet"
}

func (n *NeuralNet) Classify(vec FeatureVector) int {
	inputVec := n.network.Input()
	for i := range inputVec {
		inputVec[i] = 0
	}
	for _, v := range vec {
		inputVec[v.Index] = v.Value
	}
	n.network.PropagateForward()
	var outputClass int
	var maxOutput float64
	for j, out := range n.network.Output() {
		if out > maxOutput {
			maxOutput = out
			outputClass = j
		}
	}
	return outputClass
}

func (n *NeuralNet) train(vecs []FeatureVector, classes []int, cancel <-chan struct{}) {
	n.network.Randomize()
	for {
		log.Printf("Current results: %s", n.rightCounts(vecs, classes))
		perm := rand.Perm(len(vecs))
		for _, x := range perm {
			story := vecs[x]
			class := classes[x]
			n.sgdStepStory(story, class)
			select {
			case <-cancel:
				return
			default:
			}
		}
	}
}

func (n *NeuralNet) sgdStepStory(f FeatureVector, class int) {
	inputVec := n.network.Input()
	downstream := n.network.DownstreamGradient()

	for i := range inputVec {
		inputVec[i] = 0
	}
	for _, v := range f {
		inputVec[v.Index] = v.Value
	}

	n.network.PropagateForward()

	expected := make([]float64, len(downstream))
	expected[class] = 1

	costFunc := neuralnet.MeanSquaredCost{}
	costFunc.Deriv(n.network, expected, downstream)

	n.network.PropagateBackward(false)
	n.network.StepGradient(-n.trainConfig.StepSize)
}

func (n *NeuralNet) rightCounts(vecs []FeatureVector, classes []int) string {
	rightMap := make([]int, len(n.network.Output()))
	totalMap := make([]int, len(n.network.Output()))
	var totalRight int
	for i, vec := range vecs {
		output := n.Classify(vec)
		if output == classes[i] {
			rightMap[classes[i]]++
			totalRight++
		}
		totalMap[classes[i]]++
	}
	resStrs := make([]string, len(rightMap))
	for i, right := range rightMap {
		resStrs[i] = fmt.Sprintf("%d/%d", right, totalMap[i])
	}
	return fmt.Sprintf("%d/%d (classes: %s)", totalRight, len(classes),
		strings.Join(resStrs, " "))
}

type neuralNetConfig struct {
	HiddenCount int
	StepSize    float64
}

func getNeuralNetConfig() (*neuralNetConfig, error) {
	count, err := strconv.Atoi(os.Getenv(NeuralNetHiddenSizeEnvVar))
	if err != nil {
		return nil, fmt.Errorf("missing %s environment variable", NeuralNetHiddenSizeEnvVar)
	}

	stepSize, err := strconv.ParseFloat(os.Getenv(NeuralNetStepSizeEnvVar), 64)
	if err != nil {
		return nil, fmt.Errorf("missing %s environment variable", NeuralNetStepSizeEnvVar)
	}

	return &neuralNetConfig{
		HiddenCount: count,
		StepSize:    stepSize,
	}, nil
}
