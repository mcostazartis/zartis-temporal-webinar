package main

import (
	expenseWorkflow "zartis-temporal-webinar/workflows/expense_poc/expense_workflow"
	"zartis-temporal-webinar/workflows/golang_workflow/common"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers() {
	builder := common.WorkflowBuilder{}
	builder.AddWorkflow(expenseWorkflow.SampleExpenseWorkflow)
	builder.AddActivity(expenseWorkflow.CreateExpenseActivity)
	builder.AddActivity(expenseWorkflow.WaitForDecisionActivity)
	builder.AddActivity(expenseWorkflow.PaymentActivity)
	builder.SetTaskQueue(expenseWorkflow.ApplicationName)
	builder.Run()
}

func main() {
	startWorkers()
}
