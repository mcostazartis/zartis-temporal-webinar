package common

import (
	"log"
	"os"
	"os/signal"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

/*
	The workflow activity definition
*/
type WorkflowActivity struct {
	Activity interface{}
	Options  activity.RegisterOptions
}

type WorkflowRegistration struct {
	Workflow interface{}
	Options  workflow.RegisterOptions
}

/*
	Workflow builder data
*/
type WorkflowBuilder struct {
	Workflows     []WorkflowRegistration
	Activities    []WorkflowActivity
	TaskQueue     string
	WorkflowReady chan bool
}

func (w *WorkflowBuilder) AddWorkflow(wf interface{}) *WorkflowBuilder {
	w.AddWorkflowWithOptions(wf, workflow.RegisterOptions{})
	return w
}

/*
	Sets the workflow interface to be run
*/
func (w *WorkflowBuilder) AddWorkflowWithOptions(wf interface{}, options workflow.RegisterOptions) *WorkflowBuilder {
	w.Workflows = append(w.Workflows, WorkflowRegistration{
		Workflow: wf,
		Options:  options,
	})
	return w
}

/*
	Register a new activity in the workflow with some default options
*/
func (w *WorkflowBuilder) AddActivity(activityInterface interface{}) *WorkflowBuilder {
	w.AddActivityWithOptions(activityInterface, activity.RegisterOptions{})
	return w
}

/*
	Register a new activity in the workflow with some options
*/
func (w *WorkflowBuilder) AddActivityWithOptions(activityInterface interface{}, options activity.RegisterOptions) *WorkflowBuilder {
	w.Activities = append(w.Activities, WorkflowActivity{
		Activity: activityInterface,
		Options:  options,
	})
	return w
}

/*
	Sets the task queue name for this workflow
*/
func (w *WorkflowBuilder) SetTaskQueue(taskQueue string) *WorkflowBuilder {
	w.TaskQueue = taskQueue
	return w
}

/*
	Run the worker in a blocking fashion. Stop the worker when interruptCh receives signal.
	Stops the worker with SIGINT or SIGTERM.
*/
func (w *WorkflowBuilder) Run() {
	w.RunWithOptions(worker.Options{})
}

/*
	Run the worker in a blocking fashion. Stop the worker when interruptCh receives signal.
	Stops the worker with SIGINT or SIGTERM.
*/
func (w *WorkflowBuilder) RunWithOptions(options worker.Options) {
	c, err := GetTemporalClient()

	if err != nil {
		log.Fatalf("Failed to init workflow %v", err.Error())
	}
	workflowWorker := worker.New(c, w.TaskQueue, options)

	for _, workflow := range w.Workflows {
		workflowWorker.RegisterWorkflowWithOptions(workflow.Workflow, workflow.Options)
	}

	for _, activity := range w.Activities {
		workflowWorker.RegisterActivityWithOptions(activity.Activity, activity.Options)
	}

	err = workflowWorker.Start()
	if err != nil {
		log.Fatalf("Unable to start worker due to %v", err)
	}

	if w.WorkflowReady != nil {
		w.WorkflowReady <- true
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal := <-quit

	workflowWorker.Stop()
	log.Printf("Server has shutdown gracefully on signal [%v]", signal)
}
