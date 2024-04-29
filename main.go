package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
)

const OREGON = "us-west-2"

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(OREGON))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	br := bedrock.NewFromConfig(cfg)
	result, err := br.ListFoundationModels(context.TODO(), &bedrock.ListFoundationModelsInput{})
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	if len(result.ModelSummaries) == 0 {
		fmt.Println("There are no foundation models.")
	}
	for _, modelSummary := range result.ModelSummaries {
		fmt.Println(*modelSummary.ModelId)
	}
}
