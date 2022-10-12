package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/jimmykodes/gommand"
	"github.com/jimmykodes/gommand/flags"
)

var (
	ErrEmptyInput = errors.New("response cannot be empty")
)

func init() {
	rootCmd.Flags().AddFlagSet(Flags())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &gommand.Command{
	Name:        "standup",
	Usage:       "standup",
	Description: "create standup message",
	Run: func(ctx *gommand.Context) error {
		var (
			data = struct {
				Yesterday string
				Today     string
				Blocked   bool
				OnTime    bool
			}{
				Yesterday: ctx.Flags().String("yesterday"),
				Today:     ctx.Flags().String("today"),
				Blocked:   ctx.Flags().Bool("blocked"),
				OnTime:    ctx.Flags().Bool("on-time"),
			}
			qMap = map[string]*survey.Question{
				"yesterday": {
					Name:   "yesterday",
					Prompt: &survey.Input{Message: "What did you do yesterday?"},
					Validate: func(ans interface{}) error {
						if ans.(string) == "" {
							return ErrEmptyInput
						}
						return nil
					},
				},
				"today": {
					Name:   "today",
					Prompt: &survey.Input{Message: "What are you working on today?"},
					Validate: func(ans interface{}) error {
						if ans.(string) == "" {
							return ErrEmptyInput
						}
						return nil
					},
				},
				"blocked": {
					Name:   "blocked",
					Prompt: &survey.Confirm{Message: "Are you blocked?"},
				},
				"on-time": {
					Name:   "onTime",
					Prompt: &survey.Confirm{Message: "Are you on time?"},
				},
			}
		)

		var questions []*survey.Question
		if data.Yesterday == "" {
			questions = append(questions, qMap["yesterday"])
		}
		if data.Today == "" {
			questions = append(questions, qMap["today"])
		}
		if len(questions) > 0 {
			questions = append(questions, qMap["blocked"], qMap["on-time"])
			if err := survey.Ask(questions, &data); err != nil {
				fmt.Println("error:", err)
				return err
			}
		}

		var out io.WriteCloser
		if output := ctx.Flags().String("output"); output == "" {
			out = os.Stdout
		} else {
			f, err := os.Create(output)
			if err != nil {
				fmt.Println("error creating file:", err)
				return err
			}
			out = f
		}
		defer out.Close()

		for _, item := range [][2]string{
			{":yesterday:", data.Yesterday},
			{":today:", data.Today},
			{":road-block:", stringify(data.Blocked)},
			{":on-time:", stringify(data.OnTime)},
		} {
			fmt.Fprintln(out, item[0], item[1])
		}

		return nil
	},
}

func Flags() *flags.FlagSet {
	fs := flags.NewFlagSet(flags.WithNoEnv())

	fs.StringS("yesterday", 'y', "", "what you did yesterday")
	fs.StringS("today", 't', "", "what you are doing today")
	fs.BoolS("blocked", 'b', false, "whether you are blocked")
	fs.BoolS("on-time", 'o', false, "whether you are on time")
	fs.StringS("output", 'O', "", "file to write output to (stdout if empty)")

	return fs
}

func stringify(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
