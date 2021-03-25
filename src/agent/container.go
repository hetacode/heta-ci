package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	utilsuio "github.com/hetacode/heta-ci/agent/utils/io"

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

const (
	ScriptsDir  = "/scripts"
	PipelineDir = "/pipeline"
	JobDir      = "/job"
)

// NewContainer pull docker image, create container and run it
func NewContainer(image string, scriptsDir, pipelineDir string) *Container {
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
			Cmd:          []string{"/bin/sh"},
			WorkingDir:   JobDir,
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: pipelineDir,
					Target: PipelineDir,
				},
				{
					Type:   mount.TypeBind,
					Source: scriptsDir,
					Target: ScriptsDir,
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

// ExecuteScript inside container
// Script is lying on the host directory which is mounted via volume
func (c *Container) ExecuteScript(scriptName string, logCh chan string) error {
	scriptPath := ScriptsDir + "/" + scriptName

	config := types.ExecConfig{
		Detach:       false,
		Tty:          true,
		AttachStdout: true,
		Cmd:          []string{"/bin/sh", "-e", scriptPath},
	}
	containerExecCreate, _ := c.client.ContainerExecCreate(context.Background(), c.piplineContainerID, config)
	r, _ := c.client.ContainerExecAttach(context.Background(), containerExecCreate.ID, types.ExecStartCheck{Detach: false})
	l, _ := utilsuio.ReadWithTimeout(r.Reader, time.Second*1)

	insp, _ := c.client.ContainerExecInspect(context.Background(), containerExecCreate.ID)

	result := string(l)
	logCh <- result

	if insp.ExitCode != 0 {
		return fmt.Errorf("process completed with exit code: %d", insp.ExitCode)
	}

	return nil
}

// CreateDir inside container
func (c *Container) CreateDir(path string) error {
	attached, _ := c.client.ContainerAttach(context.Background(), c.piplineContainerID, types.ContainerAttachOptions{Stream: true, Stdin: true})
	defer attached.Close()
	attached.Conn.Write([]byte(fmt.Sprintf("mkdir %s\n", path)))

	return nil
}

// Dispose container resources
// Should be invoke with defer statement
func (c *Container) Dispose() {
	c.client.ContainerRemove(context.Background(), c.piplineContainerID, types.ContainerRemoveOptions{Force: true})

}
