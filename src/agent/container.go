package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

const ContainerName = "hetaci-agent-job"

type Container struct {
	client             *client.Client
	piplineContainerID string
}

// NewContainer pull docker image, create container and run it
func NewContainer(image string) *Container {
	ctx := context.Background()
	client, err := client.NewClientWithOpts()
	if err != nil {
		log.Printf("docker init err: %s", err)
		return nil
	}

	ping, _ := client.Ping(ctx)
	log.Printf("docker ping: %v", ping.APIVersion)

	pullReader, err := client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		log.Printf("docker pull err: %s", err)
		return nil
	}

	pullReaderBytes, _ := ioutil.ReadAll(pullReader)
	log.Print(string(pullReaderBytes))
	pwd, _ := os.Getwd()
	containerRes, err := client.ContainerCreate(
		ctx,
		&container.Config{
			Image:        image,
			Tty:          true,
			OpenStdin:    true,
			AttachStdout: true,
			Cmd:          []string{"/bin/bash"},
			WorkingDir:   "/pipeline",
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: pwd + "/tests", // TODO: from constructor parameter
					Target: "/pipeline",
				},
			},
		},
		nil, nil, ContainerName,
	)

	if err != nil {
		log.Printf("docker create container err: %s", err)
		return nil
	}

	log.Printf("container is created | id: %s | name: %s\n", containerRes.ID, ContainerName)

	if err := client.ContainerStart(ctx, containerRes.ID, types.ContainerStartOptions{}); err != nil {
		log.Printf("docker start err: %s", err)
		return nil
	}

	log.Print("container is running")

	c := &Container{
		client:             client,
		piplineContainerID: containerRes.ID,
	}

	return c
}

// Dispose container resources
// Should be invoke with defer statement
func (c *Container) Dispose() {
	c.client.ContainerRemove(context.Background(), c.piplineContainerID, types.ContainerRemoveOptions{Force: true})

}
