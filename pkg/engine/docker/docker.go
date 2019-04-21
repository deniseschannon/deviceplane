package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/strslice"

	"github.com/deviceplane/deviceplane/pkg/engine"
	"github.com/deviceplane/deviceplane/pkg/spec"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var _ engine.Engine = &Engine{}

type Engine struct {
	client *client.Client
}

func NewEngine() (*Engine, error) {
	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &Engine{
		client: client,
	}, nil
}

func (e *Engine) Create(ctx context.Context, name string, s spec.Service) (string, error) {
	resp, err := e.client.ContainerCreate(ctx, &container.Config{
		Image:      s.Image,
		Entrypoint: strslice.StrSlice(s.Entrypoint),
		Cmd:        s.Command,
		Labels:     s.Labels,
	}, nil, nil, name)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (e *Engine) Start(ctx context.Context, id string) error {
	if err := e.client.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
		// TODO
		if strings.Contains(err.Error(), "No such container") {
			return engine.ErrInstanceNotFound
		}
		return err
	}
	return nil
}

func (e *Engine) List(ctx context.Context, keyFilters map[string]bool, keyAndValueFilters map[string]string, all bool) ([]engine.Instance, error) {
	args := filters.NewArgs()
	for k := range keyFilters {
		args.Add("label", k)
	}
	for k, v := range keyAndValueFilters {
		args.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	containers, err := e.client.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
		All:     all,
	})
	if err != nil {
		return nil, err
	}

	var instances []engine.Instance
	for _, container := range containers {
		instances = append(instances, convert(container))
	}

	return instances, nil
}

func (e *Engine) Stop(ctx context.Context, id string) error {
	if err := e.client.ContainerStop(ctx, id, nil); err != nil {
		// TODO
		if strings.Contains(err.Error(), "No such container") {
			return engine.ErrInstanceNotFound
		}
		return engine.ErrInstanceNotFound
	}
	return nil
}

func (e *Engine) Remove(ctx context.Context, id string) error {
	if err := e.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{}); err != nil {
		// TODO
		if strings.Contains(err.Error(), "No such container") {
			return engine.ErrInstanceNotFound
		}
		return engine.ErrInstanceNotFound
	}
	return nil
}

func convert(c types.Container) engine.Instance {
	return engine.Instance{
		ID:     c.ID,
		Labels: c.Labels,
		// TODO
		Running: c.State == "running",
	}
}
