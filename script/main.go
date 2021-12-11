package main

import (
	"fmt"
	"log"

	"github.com/ktr0731/go-fuzzyfinder/scoring"
)

func main() {
	//usr, err := getUserStats("jackiebrewster307")
	//if err != nil {
	//log.Fatal(err)
	//}

	//score, _ := scoring.Calculate("hello world goodbye", "hello W0rld")
	//fmt.Println(score)
	//fmt.Printf("%+v\n", usr)

	lookup := []string{"hallo worldo", "hello w0rldd", "g0odbai world", "hola world!"}

	//matches := matching.FindAll("hello", lookup)

	query := "hello world"

	for _, match := range lookup {
		score, _ := scoring.Calculate(match, query)
		fmt.Printf("%s %d\n", match, score)
	}

	spamQuery := "author:vilmibm is:closed comments:0"
	spamIssues, err := spam.issueSearchQuery("cli", "cli", spamQuery)
	if err != nil {
		log.Fatal(err)
	}

	for _, issue := range spamIssues {
		fmt.Printf("%+v\n", issue)
	}
}
