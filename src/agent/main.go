package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"time"

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
			Image:     "ubuntu:20.10",
			Tty:       true,
			OpenStdin: true,
			Cmd:       []string{"uname", "-a"},
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
	go func() {
		logOutReader, err := client.ContainerLogs(ctx, containerRes.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
		if err != nil {
			fmt.Printf("docker logs reading err: %s", err)
			return
		}
		scan := bufio.NewScanner(logOutReader)
		for scan.Scan() {

			fmt.Println(scan.Text())
		}
	}()

	time.Sleep(10 * time.Second)
}
