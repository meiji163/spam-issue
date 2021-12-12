package main

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/trees"
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

	classificationData, err := base.ParseCSVToInstances("cli-cli_data.csv", true)
	if err != nil {
		panic(err)
	}

	decTree := trees.NewDecisionTreeClassifier("entropy", 8, []int64{0, 1})
	err = readGob("cli-cli_model.gob", decTree)
	if err != nil {
		panic(err)
	}

	//err = decTree.Fit(classificationData)
	//if err != nil {
	//	panic(err)
	//}

	acc, err := decTree.Evaluate(classificationData)
	fmt.Printf("Accuracy: %.3f\n", acc)

	//writeGob("cli-cli_model.gob", *decTree)
}
