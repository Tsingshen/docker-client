package prune

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
)

func Prune(ctx context.Context, cli *dockerClient.Client, pruneDocker, pruneImageDangling bool, timeBeforeDays int64, pruneDockerState string) error {

	timeBeforeDaysTime := time.Now().Add(-time.Hour * 24 * time.Duration(timeBeforeDays)).Unix()

	// prune docker
	if pruneDocker {
		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
			All: true,
		})
		if err != nil {
			return err
		}

		for _, c := range containers {
			var cName string
			if len(c.Names) > 0 {
				cName = strings.Split(c.Names[0], "/")[1]
			}
			if c.State == pruneDockerState && timeBeforeDaysTime > c.Created {
				inspect, err1 := cli.ContainerInspect(ctx, c.ID)
				if err1 != nil {
					return err1
				}

				containerTime := time.Unix(c.Created, 0).Format(time.RFC3339)
				if !strings.HasPrefix(inspect.Name, "/k8s_") {
					log.Printf(`Container "%s" , state = "%s", create at "%s" , will be prune`, cName, c.State, containerTime)
					err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{})
					if err != nil {
						return err
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
		return err
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
				return err
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

	return nil
}
