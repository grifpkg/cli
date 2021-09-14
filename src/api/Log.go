package api

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"time"
)

type Message struct {
	Value string
	Color interface{}
}

type LogLevel int
const (
	Progress LogLevel = iota
	Info
	Warn
	Success
	Question
)

var currentSpinner *spinner.Spinner = spinner.New(spinner.CharSets[40],0)

func Ask(questions []*survey.Question, answers interface{}){
	LogTool(Question, []Message{}, questions, answers)
}

func Log(level LogLevel, messages []Message){
	LogTool(level, messages, nil, nil)
}

func LogOne(level LogLevel, message string){
	LogTool(level, []Message{
		{
			Value: message,
			Color: nil,
		},
	}, nil, nil)
}

func LogTool(level LogLevel, messages []Message, questions interface{}, answers interface{}) {
	if level== Progress {
		if currentSpinner.Delay==0 {
			currentSpinner = spinner.New(spinner.CharSets[40], 50*time.Millisecond)
			currentSpinner.Prefix = "  "
			err := currentSpinner.Color("fgHiGreen","bold")
			if err != nil {
				// ignore, this console doesn't support ansi colors
			}
			currentSpinner.Start()
		}
		currentSpinner.Suffix = "  " + messages[0].Value
	} else {
		if currentSpinner.Delay!=0 {
			currentSpinner.Stop()
			currentSpinner =spinner.New(spinner.CharSets[40],0)
			fmt.Print("\u001b[2K\r")
		}
		if level== Question {
			if questions!=nil {
				formattedQuestions := questions.([]*survey.Question)
				err := survey.Ask(formattedQuestions, answers, survey.WithIcons(func(icons *survey.IconSet) {
					icons.Question.Text = "  ? "
					icons.Question.Format = "cyan"
					icons.SelectFocus.Text = "  • "
					icons.SelectFocus.Format = "cyan+b"
					icons.Error.Text = "  × "
					icons.Error.Format = "red+h"
					icons.Help.Text = "  i "
					icons.Help.Format = "cyan+b"
					icons.HelpInput.Text =  "  > "
					icons.HelpInput.Format = "cyan"
					icons.MarkedOption.Text = " × "
					icons.MarkedOption.Format = "cyan"
					icons.UnmarkedOption.Text = "   "
				}))
				if err != nil {
					answers = nil
					LogTool(Warn, []Message{
						{
							Value: "error while expecting user input: "+err.Error(),
							Color: nil,
						},
					}, nil, nil)
					return
				}
			} else {
				LogTool(Warn, []Message{
					{
						Value: "survey started but no questions/answers were provided",
						Color: nil,
					},
				}, nil, nil)
			}
		} else {
			prefix := " ¿ "
			c := color.New(color.FgHiWhite)
			switch level {
			case Info:
				prefix = " i "
				c = color.New(color.FgHiBlue)
			case Warn:
				prefix = " × "
				c = color.New(color.BgHiRed, color.FgHiWhite)
			case Success:
				prefix = " ✓ "
				c = color.New(color.FgHiGreen)
			}
			_, err := fmt.Fprintf(color.Output, " %s", c.SprintFunc()(prefix))
			for _, message := range messages {
				if message.Color==nil {
					_, err = fmt.Fprintf(color.Output, " %s", message.Value)
				} else {
					_, err = fmt.Fprintf(color.Output, " %s", message.Color.(*color.Color).SprintFunc()(message.Value))
				}
			}
			fmt.Print("\n")
			if err != nil {
				return
			}
		}
	}
}