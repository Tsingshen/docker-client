package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var (
	timeBeforeDays     int64
	pruneDockerState   string
	pruneImageDangling bool
	pruneDocker        bool
)

func main() {

	flag.Int64Var(&timeBeforeDays, "time-bofore-days", 0, "set coantainer and images time before days to prune")
	flag.StringVar(&pruneDockerState, "prune-container-state", "exited", "status of container to prune")
	flag.BoolVar(&pruneDocker, "prune-container", false, "switch of prune container")
	flag.BoolVar(&pruneImageDangling, "prune-dangling", true, "prune images of dangling")
	flag.Parse()

	timeBeforeDaysTime := time.Now().Add(-time.Hour * 24 * time.Duration(timeBeforeDays)).Unix()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithTimeout(time.Minute*5),
		client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// prune docker
	if pruneDocker {
		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
			All: true,
		})
		if err != nil {
			panic(err)
		}

		for _, c := range containers {
			var cName string
			if len(c.Names) > 0 {
				cName = strings.Split(c.Names[0], "/")[1]
			}
			if c.State == pruneDockerState && timeBeforeDaysTime > c.Created {
				containerTime := time.Unix(c.Created, 0).Format(time.RFC3339)
				log.Printf(`Container "%s" , state = "%s", create at "%s" , will be prune`, cName, c.State, containerTime)
				inspect, err1 := cli.ContainerInspect(ctx, c.ID)
				if err1 != nil {
					panic(err1)
				}

				if !strings.HasPrefix(inspect.Name, "k8s_") {
					err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{})
					if err != nil {
						log.Printf(`Remove container "%s" , error: %v`, cName, err)
					}
					log.Printf(`Remove container "%s" success`, cName)
				}
			}
		}

	}

	// prune image
	images, err := cli.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(
			filters.Arg("dangling", fmt.Sprintf("%t", pruneImageDangling)),
		),
	})

	if err != nil {
		panic(err)
	}

	for _, i := range images {
		var iName string
		if len(i.RepoDigests) > 0 {
			iName = strings.Split(i.RepoDigests[0], "@")[0]
		}
		if timeBeforeDaysTime > i.Created {
			imageTime := time.Unix(i.Created, 0).Format(time.RFC3339)
			log.Printf(`Image name = "%s", created at "%s", will be prune`, iName, imageTime)
			res, err := cli.ImageRemove(ctx, i.ID, types.ImageRemoveOptions{})
			if err != nil {
				log.Printf(`remove image "%s" error: %v`, iName, err)
			}
			for _, r := range res {
				if r.Untagged != "" {
					log.Printf(`Untagged: %s`, r.Untagged)
				} else {
					log.Printf(`Deleted: %s`, r.Deleted)
				}
			}
			log.Printf(`Remove image "%s" success`, iName)
		}
	}
}
