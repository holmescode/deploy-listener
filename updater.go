package main

import (
	"context"

	log "github.com/Sirupsen/logrus"

	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Updater encapsulates a Docker rolling update task
type Updater struct {
	client *client.Client
}

// NewUpdater creates a new Updater from the environment
func NewUpdater(image, tag string) (*Updater, error) {
	u := &Updater{}

	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"image": image,
		"tag":   tag,
	}).Info("Searching for matching services")
	services, err := client.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		// TODO: Make configurable
		if tag != "latest" {
			continue
		}

		spec := service.Spec
		specImage := spec.TaskTemplate.ContainerSpec.Image
		newImage := fmt.Sprintf("%s:%s", image, tag)
		if strings.HasPrefix(specImage, newImage) {
			log.WithFields(log.Fields{
				"image":      specImage,
				"service":    service.Spec.Name,
				"service_id": service.ID,
			}).Info("Found service, staring update")

			spec.TaskTemplate.ContainerSpec.Image = newImage
			response, err := client.ServiceUpdate(
				context.Background(),
				service.ID,
				service.Version,
				spec,
				types.ServiceUpdateOptions{},
			)

			if err != nil {
				log.WithFields(log.Fields{
					"service":    service.Spec.Name,
					"service_id": service.ID,
				}).Error(err)

				continue
			}

			log.WithFields(log.Fields{
				"warnings":   response.Warnings,
				"service":    service.Spec.Name,
				"service_id": service.ID,
			}).Info("Service update queued")
		}
	}

	u.client = client
	return u, nil
}
