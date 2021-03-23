package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hetacode/heta-ci/agent/structs"
	utilsio "github.com/hetacode/heta-ci/agent/utils/io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	client, err := client.NewClientWithOpts()
	if err != nil {
		fmt.Printf("docker init err: %s", err)
		return
	}

	ping, _ := client.Ping(ctx)
	fmt.Printf("docker ping: %v", ping.APIVersion)

	pullReader, err := client.ImagePull(ctx, "ubuntu:20.10", types.ImagePullOptions{})
	if err != nil {
		fmt.Printf("docker pull err: %s", err)
		return
	}

	pullReaderBytes, _ := ioutil.ReadAll(pullReader)
	fmt.Print(string(pullReaderBytes))
	pwd, _ := os.Getwd()
	containerRes, err := client.ContainerCreate(
		ctx,
		&container.Config{
			Image:        "ubuntu:20.10",
			Tty:          true,
			OpenStdin:    true,
			AttachStdout: true,
			Cmd:          []string{"/bin/bash"},
			WorkingDir:   "/data",
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: pwd + "/tests",
					Target: "/data",
				},
			},
		},
		nil, nil, ContainerName,
	)
	defer client.ContainerRemove(ctx, containerRes.ID, types.ContainerRemoveOptions{Force: true})

	if err != nil {
		fmt.Printf("docker create container err: %s", err)
		return
	}

	fmt.Printf("container is created | id: %s | name: %s\n", containerRes.ID, ContainerName)

	if err := client.ContainerStart(ctx, containerRes.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Printf("docker start err: %s", err)
		return
	}

	fmt.Println("container is running")
	scriptCommand := "/bin/bash -e %s|| echo \"Error: exit code $?\"\n"
	containerAttach, _ := client.ContainerAttach(ctx, containerRes.ID, types.ContainerAttachOptions{Stream: true, Stdin: true, Stdout: true, Stderr: true})
	defer containerAttach.Close()
	containerAttach.Conn.Write([]byte(fmt.Sprintf(scriptCommand, "/data/test_correct.sh")))
	l, _ := utilsio.ReadWithTimeout(containerAttach.Reader, time.Second*1)
	fmt.Println(string(l))

	containerAttach.Conn.Write([]byte(fmt.Sprintf(scriptCommand, "/data/test_fail.sh")))
	l, _ = utilsio.ReadWithTimeout(containerAttach.Reader, time.Second*1)
	fmt.Println(string(l))
}

func preparePipeline() *structs.Pipeline {
	pipeline := &structs.Pipeline{
		Name: "Test shell scripts in one container",
		Jobs: []structs.Job{
			{
				Name:   "Container job",
				Runner: "ubuntu:20.10",
				Tasks: []structs.Task{
					{
						Name: "Correct script",
						Command: []string{
							"echo Start",
							"cd /etc && ls -al",
							"echo End",
						},
					},
					{
						Name: "Failed script",
						Command: []string{
							"echo Start",
							"cd /etc && lt -al",
							"echo End",
						},
					},
				},
			},
		},
	}

	return pipeline
}
