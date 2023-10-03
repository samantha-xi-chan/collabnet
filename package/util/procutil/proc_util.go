package procutil

import (
	"bufio"
	"collab-net/internal/config"
	"collab-net/package/util/stl"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"sync"
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
//	//cmd := fmt.Sprintf("docker logs --follow %s", containerID)
//	logrus.Debug("cmd : ", cmd)
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

func StartProcBloRt(stdOut chan string, stdErr chan string, cb Callback, trackLog bool, cmdName string, cmdArg ...string) (funcErrCode int, procErrCode int, e error) {
	//strArr := strings.Fields(strings.TrimSpace(cmdStr))
	//log.Println(strArr)
	debugLogStdout, debugLogStderr, e := config.GetDebugLogStd()

	var wg sync.WaitGroup

	cmd := exec.Command(cmdName, cmdArg...)

	if trackLog {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println("StdoutPipe: ", err)
			return FUNC_ERROR_CODE_UNKNOW, 0, nil
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Println("StderrPipe: ", err)
			return FUNC_ERROR_CODE_UNKNOW, 0, nil
		}

		err = cmd.Start()
		if err != nil {
			log.Println("cmd.Start: ", err)
			return FUNC_ERROR_CODE_UNKNOW, 0, nil
		}
		pid := cmd.Process.Pid
		fmt.Println("Process ID:", pid)
		cb(pid)
		go func() {
			wg.Add(1)
			defer wg.Done()

			logrus.Debug("stdout start")
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				str := scanner.Text()
				stdOut <- str

				if debugLogStdout {
					log.Println("StartProcBlo out str : ", str)
				}
			}
			outErr := stdout.Close()
			if outErr != nil {
				logrus.Debug("outerr := stdout.Close() err:", err)
			}
			logrus.Debug("stdout end")
		}()

		go func() {
			wg.Add(1)
			defer wg.Done()

			logrus.Debug("stderr start")
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				str := scanner.Text()
				stdErr <- str

				if debugLogStderr {
					log.Println("StartProcBlo err str : ", str)
				}
			}
			errErr := stderr.Close()
			if errErr != nil {
				logrus.Debug("outerr := stdout.Close() err:", err)
			}
			logrus.Debug("stderr end")
		}()
	} else {
		err := cmd.Start()
		if err != nil {
			log.Println("cmd.Start: ", err)
			return FUNC_ERROR_CODE_UNKNOW, 0, nil
		}
		pid := cmd.Process.Pid
		fmt.Println("Process ID:", pid)
		cb(pid)
	}

	err := cmd.Wait()
	if err != nil {
		log.Println("cmd.Wait err:  ", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			log.Println("exec.ExitError ExitCode: ", exitCode)
			return 0, exitCode, nil
		} else {
			log.Println("exec.ExitError: ", err)
			return 0, 0, errors.Wrap(err, "")
		}
	}

	wg.Wait()
	log.Println("cmd.Wait end ok")
	return 0, 0, nil
}

func StartProcBlo(stdOut chan string, stdErr chan string, cb Callback, trackLog bool, cmdName string, cmdArg ...string) (funcErrCode int, procErrCode int, e error) {

	debugLogStdout, _, e := config.GetDebugLogStd()

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
			if debugLogStdout {
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
