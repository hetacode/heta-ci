package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/hetacode/heta-ci/agent/errors"
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
	ContainerScriptsDir   = "/scripts"
	ContainerArtifactsDir = "/artifacts"
	ContainerJobDir       = "/job"
	ContainerTasksDir     = "/tasks"
	ContainerCodeDir      = "/code"
)

// NewContainer pull docker image, create container and run it
func NewContainer(image string, scriptsAgentDir, artifactsAgentDir, codeAgentDir string) *Container {
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
			WorkingDir:   ContainerJobDir,
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: artifactsAgentDir,
					Target: ContainerArtifactsDir,
				},
				{
					Type:   mount.TypeBind,
					Source: scriptsAgentDir,
					Target: ContainerScriptsDir,
				},
				{
					Type:   mount.TypeBind,
					Source: codeAgentDir,
					Target: ContainerCodeDir,
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
func (c *Container) ExecuteScript(scriptName string, environments []string) (string, error) {
	scriptPath := ContainerScriptsDir + "/" + scriptName

	config := types.ExecConfig{
		Detach:       false,
		Tty:          true,
		AttachStdout: true,
		Env:          environments,
		Cmd:          []string{"/bin/sh", "-e", scriptPath},
	}
	containerExecCreate, _ := c.client.ContainerExecCreate(context.Background(), c.piplineContainerID, config)
	r, _ := c.client.ContainerExecAttach(context.Background(), containerExecCreate.ID, types.ExecStartCheck{Detach: false})
	l, _ := utilsuio.ReadWithTimeout(r.Reader, time.Second*1)

	insp, _ := c.client.ContainerExecInspect(context.Background(), containerExecCreate.ID)

	result := string(l)

	if insp.ExitCode != 0 {
		return result, &errors.ContainerError{ErrorCode: insp.ExitCode, Message: fmt.Sprintf("process completed with exit code: %d", insp.ExitCode)}
	}

	return result, nil
}

// CreateDir inside container
func (c *Container) CreateDir(path string) error {
	attached, _ := c.client.ContainerAttach(context.Background(), c.piplineContainerID, types.ContainerAttachOptions{Stream: true, Stdin: true})
	defer attached.Close()
	attached.Conn.Write([]byte(fmt.Sprintf("mkdir %s\n", path)))

	return nil
}

// RemoveDir inside container
func (c *Container) RemoveDir(path string) error {
	attached, _ := c.client.ContainerAttach(context.Background(), c.piplineContainerID, types.ContainerAttachOptions{Stream: true, Stdin: true})
	defer attached.Close()
	attached.Conn.Write([]byte(fmt.Sprintf("rm -rf %s\n", path)))

	return nil
}

// Dispose container resources
// Should be invoke with defer statement
func (c *Container) Dispose() {
	c.client.ContainerRemove(context.Background(), c.piplineContainerID, types.ContainerRemoveOptions{Force: true})

}
