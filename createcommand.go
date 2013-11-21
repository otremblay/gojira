package main

import (
	"strings"
)

var createCommand CreateCommand

func init() {
	parser.AddCommand("create", "Create a task or story", "Allows you to create a subtask or story for the given options", &createCommand)
}

type CreateCommand struct {
	Parent   string `short:"p" long:"parent" description:"Parent of the task you're creating."`
	Estimate string `short:"e" long:"estimate" description:"Your original estimate of the story"`
}

func (cc *CreateCommand) Execute(args []string) error {
	jc := NewJiraClient(options)
	if options.Project == "" {
		return &CommandError{"gojira -p flag is required for this."}
	}
	opts := &newTaskOptions{}
	opts.OriginalEstimate = cc.Estimate
	opts.Summary = strings.Join(args[1:], " ")
	opts.TaskType = args[0]
	if cc.Parent != "" {
		iss, err := jc.GetIssue(cc.Parent)
		opts.Parent = iss
		if err != nil {
			return err
		}
	}
	err := jc.CreateTask(options.Project, opts)
	if err != nil {
		return err
	}

	if len(args) > 3 {

	} else {
		return &CommandError{"Not enough arguments"}
	}
	return nil

}

type newTaskOptions struct {
	TaskType         string
	Summary          string
	OriginalEstimate string
	Parent           *Issue
}
