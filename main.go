package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	OREGON     = "us-west-2"
	errorColor = color.FgRed
)

func main() {
	os.Exit(run(initApp()))
}

func run(app *cli.App) int {
	return msg(app.Run(os.Args))
}

func msg(err error) int {
	if err != nil {
		red := color.New(errorColor).FprintfFunc()
		red(os.Stderr, "%s: %v\n", os.Args[0], err)
		return 1
	}
	return 0
}

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = "tsuji"
	app.Usage = "an English learning tool utilizing Generative AI."
	app.Version = "0.0.1"
	app.Commands = []*cli.Command{
		{
			Name:  "list",
			Usage: "list available models",
			Action: func(ctx *cli.Context) error {
				cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(OREGON))
				if err != nil {
					return err
				}

				br := bedrock.NewFromConfig(cfg)
				result, err := br.ListFoundationModels(context.TODO(), &bedrock.ListFoundationModelsInput{})
				if err != nil {
					return err
				}
				if len(result.ModelSummaries) == 0 {
					fmt.Println("There are no foundation models.")
				}
				for _, modelSummary := range result.ModelSummaries {
					fmt.Println(*modelSummary.ModelId)
				}
				return nil
			},
		},
		{
			Name:  "example",
			Usage: "example output",
			Action: func(ctx *cli.Context) error {
				cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(OREGON))
				if err != nil {
					return err
				}

				payload := Claude3Request{
					AmthoropicVersion: "bedrock-2023-05-31",
					MaxTokens:         512,
					Temperature:       0.5,
					Messages: []ClaudeMessage{
						{
							Role: "user",
							Content: []ClaudeMessageContent{
								{
									Type: "text",
									Text: "Hello",
								},
							},
						},
					},
				}
				body, _ := json.Marshal(payload)
				fmt.Println(string(body))

				brc := bedrockruntime.NewFromConfig(cfg)
				res, err := brc.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
					ModelId:     aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
					ContentType: aws.String("application/json"),
					Body:        body,
				})
				if err != nil {
					return err
				}
				fmt.Println(string(res.Body))

				return nil
			},
		},
	}
	app.Action = appRun
	return app
}

func appRun(c *cli.Context) error {
	args := c.Args()
	if !args.Present() {
		cli.ShowAppHelp(c)
	}
	return nil
}

type Claude3Request struct {
	AmthoropicVersion string          `json:"anthropic_version"`
	MaxTokens         int             `json:"max_tokens"`
	System            string          `json:"system,omitempty"`
	Messages          []ClaudeMessage `json:"messages"`
	Temperature       float64         `json:"temperature"`
	TopP              float64         `json:"top_p,omitempty"`
}

type ClaudeMessage struct {
	Role    string                 `json:"role"`
	Content []ClaudeMessageContent `json:"content"`
}

type ClaudeMessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
