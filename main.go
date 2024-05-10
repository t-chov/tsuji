package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	OREGON        = "us-west-2"
	errorColor    = color.FgRed
	SYSTEM_PROMPT = "As a teacher specialized in English learning, please respond in a way that improves the English ability of the learners."
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
		// TODO:
		// 3. 英語学習用に追い込む
		// 4. テスト
		{
			Name:  "translate",
			Usage: "translate input",
			Action: func(ctx *cli.Context) error {
				cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(OREGON))
				if err != nil {
					return err
				}

				reader := bufio.NewReader(os.Stdin)
				text, err := reader.ReadString('\n')
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
									Text: "please translate following messages to English\n" + text,
								},
							},
						},
					},
				}
				body, _ := json.Marshal(payload)

				brc := bedrockruntime.NewFromConfig(cfg)
				res, err := brc.InvokeModelWithResponseStream(context.TODO(), &bedrockruntime.InvokeModelWithResponseStreamInput{
					ModelId:     aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
					ContentType: aws.String("application/json"),
					Body:        body,
				})
				if err != nil {
					return err
				}
				if _, err := processStreamingOutput(res, func(ctx context.Context, part []byte) error {
					fmt.Print(string(part))
					return nil
				}); err != nil {
					return err
				}

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

func processStreamingOutput(
	output *bedrockruntime.InvokeModelWithResponseStreamOutput,
	handler streamingOutputHandler,
) (StreamResponse, error) {
	var combinedResult string
	resp := StreamResponse{}
	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			// for debug
			// fmt.Println("payload", string(v.Value.Bytes))
			var resp StreamResponse
			if err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp); err != nil {
				return resp, err
			}
			if resp.Delta != nil {
				handler(context.Background(), []byte(resp.Delta.Text))
				combinedResult += resp.Delta.Text
			}
			if resp.Type == "content_block_stop" {
				handler(context.Background(), []byte("\n"))
				combinedResult += "\n"
			}
		case *types.UnknownUnionMember:
			fmt.Println("unknown tag: ", v.Tag)
		default:
			fmt.Println("union is nil or unknown type")
		}
	}
	resp.Delta = &StreamResponseDelta{
		Type: "text_delta",
		Text: combinedResult,
	}
	return resp, nil
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

type StreamResponse struct {
	Type  string               `json:"type"`
	Index int                  `json:"index,omitempty"`
	Delta *StreamResponseDelta `json:"delta,omitempty"`
}

type StreamResponseDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type streamingOutputHandler func(ctx context.Context, part []byte) error
