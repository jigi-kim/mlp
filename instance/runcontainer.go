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
    Container container.Config
    Host container.HostConfig
}

func runContainer(name string, cfg *dockerConfig) {
    containerLogOptions := types.ContainerLogsOptions {
        ShowStdout: true,
        ShowStderr: true,
    }

    ctx := context.Background()

    cli, err := client.NewEnvClient()
    if err != nil {
        panic(err)
    }

    resp, err := cli.ContainerCreate(ctx, &cfg.Container, &cfg.Host, nil, name)
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

    out, err := cli.ContainerLogs(ctx, resp.ID, containerLogOptions)
    if err != nil {
        panic(err)
    }

    io.Copy(os.Stdout, out)
}

func main() {
    if len(os.Args) != 4 || (os.Args[1] != "train" && os.Args[1] != "test") {
        println("usage: runcontainer [train|test] lib data")
        return
    }

    act, lib, dat := os.Args[1], os.Args[2], os.Args[3]
    homeDir := os.Getenv("MLP_HOME")

    config := dockerConfig {
        Container: container.Config {
            Image: lib,
            Cmd: []string {"/home/ubuntu/src/start", act},
        },
        Host: container.HostConfig {
            Binds: []string {
                homeDir + "dat/" + dat + ":/home/ubuntu/dat",
                homeDir + "out/" + ":/home/ubuntu/out",
                homeDir + "src/" + ":/home/ubuntu/src",
            },
            Privileged: false,
            Runtime: "nvidia",
        },
    }

    runContainer(act, &config)
}
