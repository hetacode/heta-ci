package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	utilsio "github.com/hetacode/heta-ci/agent/utils/io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const ImageName = "ubuntu-test"

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

	containerRes, err := client.ContainerCreate(
		ctx,
		&container.Config{
			Image:        "ubuntu:20.10",
			Tty:          true,
			OpenStdin:    true,
			AttachStdout: true,
			Cmd:          []string{"/bin/sh"},
		},
		&container.HostConfig{},
		nil, nil, ImageName,
	)
	defer client.ContainerRemove(ctx, containerRes.ID, types.ContainerRemoveOptions{Force: true})

	if err != nil {
		fmt.Printf("docker create container err: %s", err)
		return
	}

	fmt.Printf("container is created | id: %s | name: %s\n", containerRes.ID, ImageName)

	if err := client.ContainerStart(ctx, containerRes.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Printf("docker start err: %s", err)
		return
	}

	fmt.Println("container is running")

	containerAttach, err := client.ContainerAttach(ctx, containerRes.ID, types.ContainerAttachOptions{Stream: true, Stdin: true, Stdout: true})
	defer containerAttach.Close()
	if err != nil {
		fmt.Printf("docker container attach err: %s", err)
		return
	}
	containerAttach.Conn.Write([]byte("echo 'test'\n"))
	l, _ := containerAttach.Reader.ReadBytes('\n')
	fmt.Println(string(l))
	l, _ = containerAttach.Reader.ReadBytes('\n')
	fmt.Println(string(l))

	containerAttach.Conn.Write([]byte("pwd\n"))
	l, _ = containerAttach.Reader.ReadBytes('\n')
	fmt.Println(string(l))
	l, _ = containerAttach.Reader.ReadBytes('\n')
	fmt.Println(string(l))

	containerAttach.Conn.Write([]byte("uname -a\n"))
	l, _ = containerAttach.Reader.ReadBytes('\n')
	fmt.Println(string(l))
	l, _ = containerAttach.Reader.ReadBytes('\n')
	fmt.Println(string(l))

	containerAttach.Conn.Write([]byte("cd /etc && ls -al\n"))
	l, _ = containerAttach.Reader.ReadBytes('\n')
	fmt.Println(string(l))

	bytes, _ := utilsio.ReadWithTimeout(containerAttach.Reader, time.Second)
	fmt.Print(string(bytes))
}
