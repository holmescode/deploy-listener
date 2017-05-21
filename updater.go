package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Updater encapsulates a Docker rolling update task
type Updater struct {
	client *client.Client
}

// NewUpdater creates a new Updater from the environment
func NewUpdater() (*Updater, error) {
	u := &Updater{}

	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}

	u.client = client
	return u, nil
}
