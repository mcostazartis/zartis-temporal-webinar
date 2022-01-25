package common

import (
	"context"
	"log"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type StarterBuilder struct {
	Workflow     interface{}
	TaskQueue    string
	cronSchedule string
	workflowID   string
}

/*
	Set the workflow to be invoked in temporal
*/
func (s *StarterBuilder) SetWorkflow(workflow interface{}) *StarterBuilder {
	s.Workflow = workflow
	return s
}

/*
	Sets the task queue
*/
func (s *StarterBuilder) SetTaskQueue(taskQueue string) *StarterBuilder {
	s.TaskQueue = taskQueue
	return s
}

/*
	Registers a cron execution with a given ID.

	 ┌───────────── minute (0 - 59)
	 │ ┌───────────── hour (0 - 23)
	 │ │ ┌───────────── day of the month (1 - 31)
	 │ │ │ ┌───────────── month (1 - 12)
	 │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
	 │ │ │ │ │
	 │ │ │ │ │
	 * * * * *
	CronSchedule string
*/
func (s *StarterBuilder) SetCronSchedule(cronSchedule string) *StarterBuilder {
	s.cronSchedule = cronSchedule
	return s
}

/*
	Sets a static workflow id. This is mandatory for cron executions.
*/
func (s *StarterBuilder) SetWorkflowID(workflowID string) *StarterBuilder {
	s.workflowID = workflowID
	return s
}

/*
	Runs the workflow and blocks the current thread to wait for the result
*/
func (s *StarterBuilder) RunWithResult(ctx context.Context, result interface{}, args ...interface{}) error {
	we, err := s.internalRun(ctx, args...)
	if err != nil {
		return err
	}
	err = we.Get(ctx, result)
	if err == nil {
		log.Printf("Workflow was successful, result is: %v", result)
	} else {
		log.Printf("Workflow failed, result is: %v", err)
	}
	return err
}

/*
	Runs the workflow and does not wait for the result.
	Returns:
	 - workflowrun: The current workflow execution. Run workflowrun.Get(ctx, &result) to get the result
	 - error: If an error occurs
*/
func (s *StarterBuilder) Run(ctx context.Context, args ...interface{}) (client.WorkflowRun, error) {
	return s.internalRun(ctx, args...)
}

/*
	Internal method used to invoke temporal
*/
func (s *StarterBuilder) internalRun(ctx context.Context, args ...interface{}) (client.WorkflowRun, error) {
	c, err := GetTemporalClient()
	if err != nil {
		log.Fatalf("Failed to init client %v", err.Error())
	}

	s.internalProcesPreviousCron(c)

	if len(s.workflowID) == 0 {
		s.workflowID = uuid.NewString()
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:           s.workflowID,
		TaskQueue:    s.TaskQueue,
		CronSchedule: s.cronSchedule,
	}

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, s.Workflow, args...)
	if err != nil {
		log.Fatalf("Unable to execute workflow, %v", err)
		return nil, err
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	return we, err
}

func (s *StarterBuilder) internalProcesPreviousCron(c client.Client) {
	if len(s.cronSchedule) > 0 {
		if len(s.workflowID) == 0 {
			log.Fatal("If running a cron workflow, workflowID needs to be set")
		}
		err := c.TerminateWorkflow(context.Background(), s.workflowID, "", "rescheduled")
		if err == nil {
			log.Printf("Canceled previous schedule")
		} else {
			log.Printf("Failed to cancel previous schedule: %v", err)
		}
	}
}
