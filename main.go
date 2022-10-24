package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/jimmykodes/gommand"
)

var (
	ErrEmptyInput = errors.New("response cannot be empty")
)

var (
	lastResponseFile string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type Response struct {
	Yesterday string `json:"yesterday"`
	Today     string `json:"today"`
	Blocked   bool   `json:"blocked"`
	OnTime    bool   `json:"on_time"`
}

var rootCmd = &gommand.Command{
	Name:        "standup",
	Usage:       "standup",
	Description: "create standup message",
	PreRun: func(_ *gommand.Context) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}
		lastResponseFile = filepath.Join(homeDir, ".standup")
		return nil
	},
	Run: func(ctx *gommand.Context) error {
		lastResponse, err := previousResponse()
		if err != nil {
			fmt.Println("error loading last response:", err)
			return err
		}
		questions := []*survey.Question{
			{
				Name:   "yesterday",
				Prompt: &survey.Input{Message: "What did you do yesterday?", Default: lastResponse.Today},
				Validate: func(ans interface{}) error {
					if ans.(string) == "" {
						return ErrEmptyInput
					}
					return nil
				},
			},
			{
				Name:   "today",
				Prompt: &survey.Input{Message: "What are you working on today?"},
				Validate: func(ans interface{}) error {
					if ans.(string) == "" {
						return ErrEmptyInput
					}
					return nil
				},
			},
			{
				Name:   "blocked",
				Prompt: &survey.Confirm{Message: "Are you blocked?", Default: lastResponse.Blocked},
			},
			{
				Name:   "onTime",
				Prompt: &survey.Confirm{Message: "Are you on time?", Default: lastResponse.OnTime},
			},
		}

		var data Response
		if err := survey.Ask(questions, &data); err != nil {
			fmt.Println("error:", err)
			return err
		}
		if err := saveResponse(data); err != nil {
			fmt.Println("Warning: could not save response results", err)
		}

		for _, item := range [][2]string{
			{":yesterday:", data.Yesterday},
			{":today:", data.Today},
			{":road-block:", stringify(data.Blocked)},
			{":on-time:", stringify(data.OnTime)},
		} {
			fmt.Println(item[0], item[1])
		}

		return nil
	},
}

func stringify(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func previousResponse() (Response, error) {
	var resp Response
	if _, err := os.Stat(lastResponseFile); err != nil {
		return resp, nil
	}
	f, err := os.Open(lastResponseFile)
	if err != nil {
		return resp, fmt.Errorf("open: %w", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&resp); err != nil {
		return resp, fmt.Errorf("json decode: %w", err)
	}
	return resp, nil
}

func saveResponse(resp Response) error {
	f, err := os.Create(lastResponseFile)
	if err != nil {
		return fmt.Errorf("file create: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(resp); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}
