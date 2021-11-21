package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/csw/semrelay"
)

func title(semN *semrelay.Notification) (string, error) {
	startT, err := time.Parse(time.RFC3339, semN.Pipeline.RunningAt)
	if err != nil {
		return "", err
	}
	doneT, err := time.Parse(time.RFC3339, semN.Pipeline.DoneAt)
	if err != nil {
		return "", err
	}
	mins := doneT.Sub(startT).Minutes()
	return fmt.Sprintf("Build %s for %s:%s in %.0fm",
		semN.Pipeline.Result, semN.Project.Name, semN.Revision.Branch.Name, mins), nil
}

func body(semN *semrelay.Notification) string {
	b := strings.Builder{}
	fmt.Fprintf(&b, "Commit %s: %s\n",
		semN.Revision.CommitSHA[:7], semN.Revision.CommitMessage)
	if semN.Pipeline.Result == "failed" {
		blockParts := []string{}
		for _, block := range semN.Blocks {
			jobParts := []string{}
			for _, job := range block.Jobs {
				if job.Result == "failed" {
					jobParts = append(jobParts, job.Name)
				}
			}
			if len(jobParts) > 0 {
				blockParts = append(blockParts,
					fmt.Sprintf("%s (%s)", block.Name, strings.Join(jobParts, ", ")))
			}
		}
		fmt.Fprintf(&b, "Failed in %s\n", strings.Join(blockParts, ", "))
	}
	return b.String()
}
