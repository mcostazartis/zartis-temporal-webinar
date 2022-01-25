.PHONY: clean
clean:
	rm -rf bin

.PHONY: build
build:
	go build -o bin/worker ./workflows/expense_poc/expense_workflow_worker/
	go build -o bin/dummyserver ./workflows/expense_poc/dummy_server/
