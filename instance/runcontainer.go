package main

import (
    "context"
    "io"
    "os"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)

type dockerConfig struct {
    Container *container.Config
    Host *container.HostConfig
}

func runContainer(name string, cfg *dockerConfig) {
    ctx := context.Background()

    cli, err := client.NewEnvClient()
    if err != nil {
        panic(err)
    }

    resp, err := cli.ContainerCreate(ctx, cfg.Container, cfg.Host, nil, name)
    if err != nil {
        panic(err)
    }

    err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
    if err != nil {
        panic(err)
    }

    statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
    select {
    case err := <-errCh:
        if err != nil {
            panic(err)
        }
    case <-statusCh:
    }

    out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
    if err != nil {
        panic(err)
    }

    io.Copy(os.Stdout, out)
}

func runLearner(name string, lib string, dat string) {
    workDir := "/home/ubuntu/"

    containerConfig := container.Config {
        Image: lib,
        Cmd: []string {"/home/ubuntu/src/train"},
    }

    hostConfig := container.HostConfig {
        Binds: []string {
            workDir + "out/" + ":/home/ubuntu/out",
            workDir + "src/" + ":/home/ubuntu/src",
            workDir + "dat/" + dat + ":/home/ubuntu/dat",
        },
        Privileged: false,
        Runtime: "nvidia",
    }

    config := dockerConfig {
        Container: &containerConfig,
        Host: &hostConfig,
    }

    runContainer(name, &config)
}

func main() {
    if len(os.Args) < 3 {
        println("usage: run lib data")
        return
    }

    lib := os.Args[1]
    dat := os.Args[2]

    runLearner("learner", lib, dat)
}
