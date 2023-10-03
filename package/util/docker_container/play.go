package docker_container

import (
	"collab-net/api"
	config "collab-net/internal/config"
	"collab-net/message"
	"collab-net/package/util/procutil"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
)

func WatchContainer(ctx context.Context, taskId string, containerId string, cleanContainer bool, logRt bool) (exitCode_ int, e error) {
	BUF_SIZE := 4096
	stdOut := make(chan string, BUF_SIZE)
	stdErr := make(chan string, BUF_SIZE)

	go func() {
		for msg := range stdOut {
			message.GetMsgCtl().UpdateTaskWrapper(taskId, api.TASK_STATUS_RUNNING, msg)
		}
	}()
	go func() {
		for msg := range stdErr {
			message.GetMsgCtl().UpdateTaskWrapper(taskId, api.TASK_STATUS_RUNNING, msg)
		}
	}()

	if logRt {
		funcErrCode, procErrCode, e := procutil.StartProcBloRt(stdOut, stdErr, func(pid int) {
			//log.Printf("taskId %s, pid: %d\n", taskId, pid)
		}, true, procutil.GetDockerBin(), "logs", "--follow", containerId)
		logrus.Debugf("funcErrCode: %d, procErrCode: %d, error: %s ", funcErrCode, procErrCode, e)
	} else {
		funcErrCode, procErrCode, e := procutil.StartProcBlo(stdOut, stdErr, func(pid int) {
			//log.Printf("taskId %s, pid: %d\n", taskId, pid)
		}, true, procutil.GetDockerBin(), "logs", "--follow", containerId)
		logrus.Debugf("taskId: %s, funcErrCode: %d, procErrCode: %d, error: %s ", taskId, funcErrCode, procErrCode, e)
	}

	close(stdOut)
	close(stdErr)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		e = errors.Wrap(err, "NewClientWithOpts: ")
		return
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Failed to close Docker client: %v\n", err)
		}
	}()

	containerInfo, err := cli.ContainerInspect(context.Background(), containerId)
	if err != nil {
		log.Println("ContainerInspect: e=", err)
		return 0, errors.Wrap(err, "cli.ContainerInspect: ")
	}
	exitCode := containerInfo.State.ExitCode

	if cleanContainer {
		cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		})
	}
	return exitCode, nil
}

func StartContainer(ctx context.Context,
	taskId string, callbackAddr string,
	block bool, cleanContainer bool, imageName string, cmdStringArr []string, memLimMb int64, cpuPercent int, cpuSetCpus string, containerName string,
	bindIn []api.Bind,
	bindOut []api.Bind,
) (containerId_ string, e error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		e = errors.Wrap(err, "NewClientWithOpts: ")
		return
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Failed to close Docker client: %v\n", err)
		}
	}()

	var binds = []string{config.DOCKER_PATH_BIND}

	for _, val := range bindIn {
		binds = append(binds, fmt.Sprintf("%s:%s", val.VolId, val.VolPath))
	}
	for _, val := range bindOut {
		binds = append(binds, fmt.Sprintf("%s:%s", val.VolId, val.VolPath))
	}

	log.Println("taskId: ", taskId, "  cmdStringArr: ", cmdStringArr, "  binds: ", binds)

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: imageName,
			Cmd:   cmdStringArr,
		},
		&container.HostConfig{
			Resources: container.Resources{
				Memory:     memLimMb * 1000 * 1000,     // MB
				MemorySwap: memLimMb * 1000 * 1000 * 6, // MB
				CPUPeriod:  int64(100 * 1000),
				CPUQuota:   int64(cpuPercent * 1000),
				CpusetCpus: cpuSetCpus,
			},
			LogConfig: container.LogConfig{
				Type:   "json-file",
				Config: map[string]string{"max-file": "2", "max-size": "20m"},
			},

			Binds: binds,
		},
		nil,
		nil,
		containerName,
	)
	if err != nil {
		log.Println("ContainerCreate e: ", err)
		return "", err
	}

	containerId := resp.ID
	log.Print("ContainerCreate resp: ", resp)

	err = cli.ContainerStart(
		ctx,
		containerId,
		types.ContainerStartOptions{},
	)
	if err != nil {
		log.Println("ContainerStart: ", err)
		return "", err
	}

	return containerId, nil
}

func IsContainerRunning(containerNameOrID string) (bRunning bool, e error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, errors.Wrap(err, "client.NewClientWithOpts: ")
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return false, errors.Wrap(err, "cli.ContainerList()")
	}

	for _, container := range containers {
		if container.Names[0] == "/"+containerNameOrID || container.ID == containerNameOrID {
			log.Println("container running ")
			return true, nil
		}
	}

	return false, nil
}

func StopContainer(containerId string) (e error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrap(err, "client.NewEnvClient(): ")
	}

	log.Println("trying to invoke ContainerStop , ", containerId)
	err = cli.ContainerStop(context.Background(), containerId, container.StopOptions{
		Signal:  "",
		Timeout: nil,
	})
	if err != nil {
		return errors.Wrap(err, "cli.ContainerStop(): ")
	}

	log.Println("Container Stopped, ", containerId)
	return nil
}

func ListContainer(ctx context.Context) (x []types.Container, e error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "cli.ContainerList()")
	}

	// 列出运行中的容器
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	if err != nil {
		return nil, errors.Wrap(err, "ContainerList: ")
	}

	// 打印容器信息
	for _, container := range containers {
		//fmt.Printf("容器ID：%s\n", container.ID)
		fmt.Printf("containerName：%s\n", container.Names)
		//fmt.Printf("镜像：%s\n", container.Image)
		//fmt.Printf("状态：%s\n", container.State)
		//fmt.Println("---------------")
	}

	return containers, nil
}
