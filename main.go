package main

import (
	"context"
	"docker-client/prune"
	"flag"
	"fmt"
	"time"

	"github.com/docker/docker/client"
	dockerClient "github.com/docker/docker/client"
)

var (
	timeBeforeDays     int64
	pruneDockerState   string
	pruneImageDangling bool
	pruneDocker        bool
)

func main() {

	flag.Int64Var(&timeBeforeDays, "time-bofore-days", 30, "set coantainer and images time before days to prune")
	flag.StringVar(&pruneDockerState, "prune-container-state", "exited", "status of container to prune")
	flag.BoolVar(&pruneDocker, "prune-container", false, "switch of prune container")
	flag.BoolVar(&pruneImageDangling, "prune-dangling", true, "prune images of dangling")
	flag.Parse()

	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(
		client.FromEnv,
		client.WithTimeout(time.Minute*5),
		client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	err1 := prune.Prune(ctx, cli, pruneDocker, pruneImageDangling, timeBeforeDays, pruneDockerState)
	if err != nil {
		fmt.Printf("prune get error:%v\n", err1)
	}
}
