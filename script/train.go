package main

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"

	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/ensemble"
	"github.com/sjwhitworth/golearn/evaluation"
)

func writeGob(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func readGob(filePath string, object interface{}) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}

func main() {

	rand.Seed(42)

	spamData, err := base.ParseCSVToInstances("cli-cli_data.csv", true)
	if err != nil {
		panic(err)
	}

	attrs := spamData.AllAttributes()
	fmt.Println(attrs)

	fmt.Println(spamData)
	//testData, trainData := base.InstancesTrainTestSplit(classificationData, 0.2)

	//err = readGob("cli-cli_model.gob", decTree)
	//if err != nil {
	//	panic(err)
	//}

	tree := ensemble.NewRandomForest(71, 9)

	err = tree.Fit(spamData)
	if err != nil {
		panic(err)
	}

	predictions, err := tree.Predict(spamData)
	cm, err := evaluation.GetConfusionMatrix(spamData, predictions)
	if err != nil {
		panic(err)
	}
	fmt.Println(evaluation.GetSummary(cm))

	//writeGob("cli-cli_model.gob", *tree)

}
