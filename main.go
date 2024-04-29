package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
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
