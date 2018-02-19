package main

import (
    "context"
    "io"
    "log"
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
        log.Fatal(err)
    }

    resp, err := cli.ContainerCreate(ctx, &cfg.Container, &cfg.Host, nil, name)
    if err != nil {
        log.Fatal(err)
    }

    err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
    if err != nil {
        log.Fatal(err)
    }

    cStat, cErr := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
    select {
    case err := <-cErr:
        if err != nil {
            log.Fatal(err)
        }
    case <-cStat:
    }

    out, err := cli.ContainerLogs(ctx, resp.ID, containerLogOptions)
    if err != nil {
        log.Fatal(err)
    }

    io.Copy(os.Stdout, out)
}

func main() {
    if len(os.Args) != 4 || (os.Args[1] != "train" && os.Args[1] != "test") {
        println("usage: runcontainer [train|test] library dataset")
        return
    }

    mod, lib, dat := os.Args[1], os.Args[2], os.Args[3]
    homeDir := os.Getenv("MLP_HOME") + "/"

    config := dockerConfig {
        Container: container.Config {
            Image: lib,
            Cmd: []string {"/home/ubuntu/src/start", mod},
        },
        Host: container.HostConfig {
            Binds: []string {
                homeDir + "src/" + ":/home/ubuntu/src",

                homeDir + "efs/datasets/" + dat + "/" + mod + ":/home/ubuntu/dataset",
                homeDir + "efs/user/" + dat + "/models" + ":/home/ubuntu/models",
                homeDir + "efs/user/" + dat + "/output" + ":/home/ubuntu/out",
            },
            Privileged: false,
            Runtime: "nvidia",
        },
    }

    runContainer(mod, &config)
}
