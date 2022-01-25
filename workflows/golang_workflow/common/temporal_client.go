package common

import (
	"log"

	"go.temporal.io/sdk/client"
)

var temporalClient client.Client

func newTemporalClient() error {

	c, err := client.NewClient(client.Options{
		HostPort:  "localhost:7233",
		Namespace: "default",
	})

	if err == nil {
		temporalClient = c
		return nil
	} else {
		log.Fatalf("Error creating temporal client: %v", err)
		return err
	}
}

func GetTemporalClient() (client.Client, error) {
	if temporalClient == nil {
		if err := newTemporalClient(); err != nil {
			return nil, err
		}
	}
	return temporalClient, nil
}
