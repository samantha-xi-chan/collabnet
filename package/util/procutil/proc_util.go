package procutil

import (
	"collab-net-v2/package/util/stl"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func KillProc(pid int32) (_killed int, e error) {
	var unvisited stl.Stack
	killed := 0

	root, err := getProcByPid(pid)
	if err != nil {
		return killed, errors.Wrap(err, "getProcByPid(pid) ")
	}

	current := root
	unvisited.Push(root)
	for !unvisited.IsEmpty() {
		top, _ := unvisited.Pop()
		current = top.(*process.Process)

		processes, _ := current.Children()
		for _, p := range processes {
			unvisited.Push(p)
		}

		if len(processes) == 0 { // leaf
			cmdLine, err := current.Cmdline()
			if err == nil {
				log.Println("going to kill: ", cmdLine)
			}
			err = current.Kill()
			if err != nil {
				log.Println("kill err: ", err)
			}
		}
	}

	return killed, nil
}

func getProcByPid(pid int32) (x *process.Process, err error) {

	processes, _ := process.Processes()
	for _, p := range processes {
		if p.Pid == pid {
			return p, nil
		}
	}

	return nil, errors.New("not found")
}

//func Run(cmd string) {
//	//cmd := fmt.Sprintf("docker logs --follow %s", containerId)
//	logrus_wrap.Debug("cmd : ", cmd)
//	strArr := strings.Fields(strings.TrimSpace(cmd))
//	process, err := os.StartProcess(GetDockerBin(), strArr, procAttr)
//	if err != nil {
//		fmt.Printf("Error %v starting process!", err) //
//		os.Exit(1)
//	}
//	log.Println("The Pid is: ", process.Pid)
//
//}

const (
	FUNC_ERROR_CODE_UNKNOW = -1
)

type Callback func(int)

func GetDockerBin() string {
	sysType := runtime.GOOS
	//log.Println("sysType: ", sysType)

	switch sysType {
	case "darwin": // todo : optimize coding style
		return "/usr/local/bin/docker"
	case "linux":
		return "/usr/bin/docker"
	}

	return ""
}

func WaitContainerLog(ctx context.Context, stdOut chan string, stdErr chan string, enableWatch *bool, containerId string) (procErrCode int, e error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return 0, errors.Wrap(err, "NewClientWithOpts: ")
	}

	sliceSize := 10 * 1024 * 1024

	go func() {
		logOptions := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Timestamps: false,
		}

		reader, err := cli.ContainerLogs(ctx, containerId, logOptions)
		if err != nil {
			log.Println("cli.ContainerLogs, err = ", err)
			return // 0, errors.Wrap(err, "cli.ContainerLogs: ")
		}
		defer reader.Close()

		logs := make(chan string)
		go func() {
			buf := make([]byte, sliceSize)
			var partialLine string
			for {
				n, err := reader.Read(buf)
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Fprintf(os.Stderr, "Error reading logs: %v\n", err)
					break
				}

				data := buf[:n]
				lines := strings.Split(partialLine+string(data), "\n")

				for i, line := range lines {
					if i < len(lines)-1 {
						logs <- line
					} else {
						partialLine = line
					}
				}
			}
			close(logs)
			log.Println("end of {for n, err := reader.Read(buf)} , containerId = ", containerId)
		}()

		SIZE_PREF := 8
		for slice := range logs {
			if *enableWatch {
				trimmedString := slice
				len := len(slice)
				//log.Println("len slice: ", len)

				if len >= SIZE_PREF {
					trimmedString = slice[SIZE_PREF:]
					//log.Println("trimmedString： ", trimmedString)
				}

				stdOut <- trimmedString
			}
		}

		log.Println("end of {for slice := range logs} , containerId = ", containerId)
		close(stdOut)
		close(stdErr)
	}()

	containerCh, errsCh := cli.ContainerWait(context.Background(), containerId, container.WaitConditionNotRunning)
	select {
	case containerResp := <-containerCh:
		fmt.Printf("containerId %s  , StatusCode: %d \n", containerId, containerResp.StatusCode)
	case err := <-errsCh:
		log.Println("err := <-errsCh", err)
	}

	containerInfo, err := cli.ContainerInspect(context.Background(), containerId)
	if err != nil {
		log.Println("ContainerInspect: ", err)
	}
	fmt.Printf("containerId %s  , State.ExitCode: %d \n", containerId, containerInfo.State.ExitCode)

	return containerInfo.State.ExitCode, nil
}

func StartProcBlo(stdOut chan string, stdErr chan string, cb Callback, trackLog bool, cmdName string, cmdArg ...string) (funcErrCode int, procErrCode int, e error) {
	cmd := exec.Command(cmdName, cmdArg...)

	if trackLog {
		// Run the command and capture its combined output
		combinedOutput, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
			return 0, 0, errors.Wrap(err, "cmd.CombinedOutput: ")
		}
		//log.Printf("Command output:\n%s\n", combinedOutput)
		lines := strings.Split(string(combinedOutput), "\n")
		for _, line := range lines {
			if false { // todo:
				log.Println(line)
			}
			stdOut <- line
		}
	}
	return 0, 0, nil
}

func RunCmd(cmdName string, cmdArg ...string) (x error) {
	_, e := exec.Command(cmdName, cmdArg...).CombinedOutput()
	return e
}
