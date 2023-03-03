package prune

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
)

type dockerPrune interface {
	prune() error
}

type Container struct {
	Client         *dockerClient.Client
	CreateTime     int64
	PruneImage     bool
	PruneContainer bool
}

func (c Container) Prune() error {

	timeBeforeDaysTime := time.Now().Add(-time.Hour * 24 * time.Duration(c.CreateTime)).Unix()

	// prune Container
	if c.PruneContainer {
		Containers, err := c.Client.ContainerList(context.Background(), types.ContainerListOptions{
			Filters: filters.NewArgs(
				filters.Arg("status", "exited"),
			),
		})
		if err != nil {
			return err
		}

		for _, container := range Containers {
			if timeBeforeDaysTime > container.Created && !strings.HasPrefix(container.Names[0], "/k8s_") {
				createdTimeUnix := time.Unix(container.Created, 0).Format(time.RFC3339)
				log.Printf(`Container "%s" state = "%s" created at "%s" will be prune`, container.Names[0], container.State, createdTimeUnix)
				err := removeContainer(c.Client, container.ID)
				if err != nil {
					return err
				}
			}
		}

	}

	// prune Image dangling
	if c.PruneImage {
		images, err := c.Client.ImageList(context.Background(), types.ImageListOptions{Filters: filters.NewArgs(filters.Arg("dangling", "true"))})
		if err != nil {
			return err
		}

		for _, i := range images {
			if timeBeforeDaysTime > i.Created {
				createdTimeUnix := time.Unix(i.Created, 0).Format(time.RFC3339)
				log.Printf(`Image name "%s" id "%s" created at "%s" will be prune`, strings.Split(i.RepoDigests[0], "@")[0], i.ID, createdTimeUnix)
				err := removeImage(c.Client, i.ID)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func removeImage(cli *dockerClient.Client, id string) error {
	_, err := cli.ImageRemove(context.Background(), id, types.ImageRemoveOptions{})
	return err
}

func removeContainer(cli *dockerClient.Client, id string) error {
	return cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{})
}
