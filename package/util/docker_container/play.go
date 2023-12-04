package docker_container

import (
	"collab-net-v2/api"
	"collab-net-v2/package/message"
	"collab-net-v2/package/util/procutil"
	"collab-net-v2/workflow/config_workflow"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
	"time"
)

func WatchContainer(ctx context.Context, taskId string, containerId string, cleanContainer bool, logRt bool) (exitCode_ int, e error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	BUF_SIZE := 4096
	stdOut := make(chan string, BUF_SIZE)
	stdErr := make(chan string, BUF_SIZE)

	isHot := true
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Child routine received signal to exit. taskId = ", taskId)
				return
			default:
				log.Println("WatchContainer Child routine is running.. taskId = ", taskId)
				time.Sleep(1 * 1000 * time.Millisecond)
				b, e := message.GetMsgCtl().GetTaskIsHot(taskId)
				if e != nil {
					log.Println("message.GetMsgCtl().GetTaskIsHot: e = ", e) // todo: error
					continue
				}
				log.Printf(" taskId: %s, isHot: %d \n", taskId, b)
				if b == api.TRUE {
					isHot = true
				} else {
					isHot = false
				}
			}
		}
	}()

	go func() { // todo： 协程泄漏 ？
		for msg := range stdOut {
			message.GetMsgCtl().UpdateTaskWrapper(taskId, api.TASK_STATUS_RUNNING, msg)
		}
		log.Println("end of range stdOut , containerId = ", containerId)
	}()
	go func() { // todo： 协程泄漏 ？
		for msg := range stdErr {
			message.GetMsgCtl().UpdateTaskWrapper(taskId, api.TASK_STATUS_RUNNING, msg)
		}
		log.Println("end of range stdErr , containerId = ", containerId)
	}()

	// logRt
	procErrCode, e := procutil.WaitContainerLog(context.Background(), stdOut, stdErr, &isHot, containerId)
	logrus.Debugf(" procErrCode: %d, error: %s ", procErrCode, e)

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

func CreateContainer(ctx context.Context,
	taskId string, callbackAddr string,
	block bool,
	imageName string, cmdStringArr []string, memLimMb int64, cpuPercent int, cpuSetCpus string, containerName string,
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

	var binds = []string{config_workflow.DOCKER_PATH_BIND}
	binds = append(binds, config_workflow.HOSTS_BIND)

	for _, val := range bindIn {
		volName := fmt.Sprintf("%s_%s", containerName, val.VolId)
		binds = append(binds, fmt.Sprintf("%s:%s", volName, val.VolPath))
	}
	for _, val := range bindOut {
		volName := fmt.Sprintf("%s_%s", containerName, val.VolId)
		binds = append(binds, fmt.Sprintf("%s:%s", volName, val.VolPath))
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
				Memory: memLimMb * 1000 * 1000, // MB
				//MemorySwap: memLimMb * 1000 * 1000 * 6, // MB
				CPUPeriod:  int64(100 * 1000),
				CPUQuota:   int64(cpuPercent * 1000),
				CpusetCpus: cpuSetCpus,
			},
			LogConfig: container.LogConfig{
				Type:   "json-file",
				Config: map[string]string{"max-file": "2", "max-size": "20m"},
			},
			ShmSize: 512 * 1024 * 1024, // in bytes

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

	return containerId, nil
}

func StartContainer(ctx context.Context, containerId string) (err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		err = errors.Wrap(err, "NewClientWithOpts: ")
		return
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Failed to close Docker client: %v\n", err)
		}
	}()

	return cli.ContainerStart(
		ctx,
		containerId,
		types.ContainerStartOptions{},
	)
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

func StopContainerByName(containerName string) (e error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrap(err, "cli.ContainerList()")
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: false,
	})
	if err != nil {
		return errors.Wrap(err, "ContainerList: ")
	}

	for _, item := range containers {
		fmt.Printf("containerName：%s\n", item.Names)
		if strings.Contains(item.Names[0], containerName) {
			log.Println("trying to invoke ContainerStop: ", containerName, " ", item.ID)
			err = cli.ContainerStop(context.Background(), item.ID, container.StopOptions{
				Signal:  "",
				Timeout: nil,
			})
			if err != nil {
				return errors.Wrap(err, "cli.ContainerStop(): ")
			} else {
				log.Println(" ContainerStop OK:  ", containerName, " ", item.ID)
			}
		}

	} // end  of for

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
