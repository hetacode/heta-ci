package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	utilsio "github.com/hetacode/heta-ci/agent/utils/io"
)

const ContainerName = "hetaci-agent-job"

type Container struct {
	client             *client.Client
	piplineContainerID string
}

// NewContainer pull docker image, create container and run it
func NewContainer(image string, pipelineTempDir string) *Container {
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
					Source: pipelineTempDir, // TODO: from constructor parameter
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

func (c *Container) ExecuteScript(scriptName string, logCh chan string) error {
	scriptCommand := "/bin/bash -e /pipeline/scripts/%s || echo \"Error: exit code $?\"\n"

	containerAttach, _ := c.client.ContainerAttach(context.Background(), c.piplineContainerID, types.ContainerAttachOptions{Stream: true, Stdin: true, Stdout: true, Stderr: true})
	defer containerAttach.Close()
	containerAttach.Conn.Write([]byte(fmt.Sprintf(scriptCommand, scriptName)))
	l, _ := utilsio.ReadWithTimeout(containerAttach.Reader, time.Second*1)
	result := string(l)
	logCh <- result

	if len(result) > 30 {
		end := result[:len(result)-30]
		log.Printf("\nend: %s\n", end)
		if strings.Contains(end, "Error: exit code") && !strings.Contains(end, "echo") {
			return fmt.Errorf("script %s execute with error", scriptName)
		}
	}

	return nil
}

// Dispose container resources
// Should be invoke with defer statement
func (c *Container) Dispose() {
	c.client.ContainerRemove(context.Background(), c.piplineContainerID, types.ContainerRemoveOptions{Force: true})

}
