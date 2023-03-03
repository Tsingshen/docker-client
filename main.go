package main

import (
	"docker-client/prune"
	"flag"
	"fmt"
	"time"

	"github.com/docker/docker/client"
	dockerClient "github.com/docker/docker/client"
)

var (
	timeBeforeDays int64
)

func main() {

	flag.Int64Var(&timeBeforeDays, "time-bofore-days", 30, "set coantainer and images time before days to prune")
	flag.Parse()

	cli, err := dockerClient.NewClientWithOpts(
		client.FromEnv,
		client.WithTimeout(time.Minute*5),
		client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	container := prune.Container{
		Client:         cli,
		CreateTime:     timeBeforeDays,
		PruneImage:     true,
		PruneContainer: true,
	}

	err1 := container.Prune()
	if err1 != nil {
		fmt.Printf("prune get error:%v\n", err1)
	}
}
