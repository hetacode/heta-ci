package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/docker/docker/api/types"
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
}
