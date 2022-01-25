package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	expenseWorkflow "zartis-temporal-webinar/workflows/expense_poc/expense_workflow"
	"zartis-temporal-webinar/workflows/golang_workflow/common"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
)

type expenseState string

const (
	created   expenseState = "CREATED"
	approved               = "APPROVED"
	rejected               = "REJECTED"
	completed              = "COMPLETED"
)

const tpl = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Compliance POC</title>

        <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-wEmeIV1mKuiNpC+IOBjI7aAzPcEZeedi5yW5f2yOq55WWLwNGmvvx4Um1vskeMj0" crossorigin="anonymous">

        <style>
            .logo, .logo .st0{
                width: 130px;
                height: 70px;
                margin: auto;
                fill: #d11a17;
            }

            .spacers {
                padding: 40px 0 20px 0;
            }
        </style>

        <meta name="viewport" content="width=device-width, initial-scale=1">
	</head>
	<body>
	    <!-- Option 1: Bootstrap Bundle with Popper -->
        <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0/dist/js/bootstrap.bundle.min.js" integrity="sha384-p34f1UUtsS3wqzfto5wAAmdvj+osOnFyQFpp4Ua3gs/ZVWx6oOypYoCJhGGScy+8" crossorigin="anonymous"></script>

        <div class="container">
            <div class="row">
                <h1>
					<img src="https://www.zartis.com/wp-content/uploads/2021/11/zartis-logo-black.svg" />
				</h1>
            </div>

            <div class="spacers">
                <a class="btn btn-primary btn-lg" href="/list">HOME</a>
                <a class="btn btn-secondary btn-lg" href="/new">CREATE NEW COMPLIANCE REQUEST</a>
            </div>

            <div class="row spacers">
                <h3>All compliance requests:</h3>
            </div>

            <div class="row">
                <table class="table">
                    <tr>
                        <th>Request ID</th>
                        <th>Status</th>
                        <th>Action</th>
                    </tr>
                    {{range .Items}}
                    <tr>
                        <td>{{ .Id }}</td>
                        <td>{{ .Status }}</td>
                        <td>
                            {{if eq .Status "CREATED"}}
                                <a class="btn btn-success" href="/action?type=approve&id={{ .Id}}">APPROVE</a>
                                <a class="btn btn-danger" href="/action?type=reject&id={{ .Id}}">REJECT</a>
                            {{end}}
                        </td>
                    </tr>
                    {{else}}
                    <tr>
                        <td colspan="3"><strong>No rows</strong></td>
                    </th>
                    {{end}}
                </table>
            </div>
        </div>
	</body>
</html>`

// use memory store for this dummy server
var allExpense = make(map[string]expenseState)

var tokenMap = make(map[string][]byte)

var workflowClient client.Client

type Expense struct {
	Id     string
	Status expenseState
}

func main() {
	var err error
	workflowClient, err = common.GetTemporalClient()
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting dummy server...")
	http.HandleFunc("/", listHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/action", actionHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/registerCallback", callbackHandler)
	http.HandleFunc("/new", newHandler)
	http.ListenAndServe(":8099", nil)
}

func startWorkflow(expenseID string) {
	builder := common.StarterBuilder{}
	builder.SetWorkflow(expenseWorkflow.SampleExpenseWorkflow)
	builder.SetWorkflowID("expense_" + uuid.New())
	builder.SetTaskQueue(expenseWorkflow.ApplicationName)

	builder.Run(context.Background(), expenseID)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	keys := []string{}
	for k := range allExpense {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	data := struct {
		Items []Expense
	}{
		Items: make([]Expense, 0),
	}

	for _, id := range keys {
		data.Items = append(data.Items, Expense{
			Id:     id,
			Status: allExpense[id],
		})

	}
	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func actionHandler(w http.ResponseWriter, r *http.Request) {
	isAPICall := r.URL.Query().Get("is_api_call") == "true"
	id := r.URL.Query().Get("id")
	oldState, ok := allExpense[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}
	actionType := r.URL.Query().Get("type")
	switch actionType {
	case "approve":
		allExpense[id] = approved
	case "reject":
		allExpense[id] = rejected
	case "payment":
		allExpense[id] = completed
	}

	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		http.Redirect(w, r, "/list", 302)
	}

	if oldState == created && (allExpense[id] == approved || allExpense[id] == rejected) {
		// report state change
		notifyExpenseStateChange(id, string(allExpense[id]))
	}

	fmt.Printf("Set state for %s from %s to %s.\n", id, oldState, allExpense[id])
	return
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	isAPICall := r.URL.Query().Get("is_api_call") == "true"
	id := r.URL.Query().Get("id")
	_, ok := allExpense[id]
	if ok {
		fmt.Fprint(w, "ERROR:ID_ALREADY_EXISTS")
		return
	}

	allExpense[id] = created
	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		listHandler(w, r)
	}
	fmt.Printf("Created new expense id:%s.\n", id)
	return
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	state, ok := allExpense[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}

	fmt.Fprint(w, state)
	fmt.Printf("Checking status for %s: %s\n", id, state)
	return
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	currState, ok := allExpense[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}
	if currState != created {
		fmt.Fprint(w, "ERROR:INVALID_STATE")
		return
	}

	err := r.ParseForm()
	if err != nil {
		// Handle error here via logging and then return
		fmt.Fprint(w, "ERROR:INVALID_FORM_DATA")
		return
	}

	taskToken := r.PostFormValue("task_token")
	fmt.Printf("Registered callback for ID=%s, token=%s\n", id, taskToken)
	tokenMap[id] = []byte(taskToken)
	fmt.Fprint(w, "SUCCEED")
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	isAPICall := r.URL.Query().Get("is_api_call") == "true"

	id := uuid.New()
	startWorkflow(id)

	time.Sleep(1 * time.Second)

	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		http.Redirect(w, r, "/list", 302)
	}

	fmt.Printf("Created new compliance request:%s.\n", id)
	return
}

func notifyExpenseStateChange(id, state string) {
	token, ok := tokenMap[id]
	if !ok {
		fmt.Printf("Invalid id:%s\n", id)
		return
	}
	err := workflowClient.CompleteActivity(context.Background(), token, state, nil)
	if err != nil {
		fmt.Printf("Failed to complete activity with error: %+v\n", err)
	} else {
		fmt.Printf("Successfully complete activity: %s\n", token)
	}
}
