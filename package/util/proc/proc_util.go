package proc

import (
	"bufio"
	"collab-net/package/util/stl"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
	"log"
	"os/exec"
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

const (
	FUNC_ERROR_CODE_UNKNOW = -1
)

type Callback func(int)

func StartProcBlo(stdOut chan string, stdErr chan string, cb Callback, trackLog bool, cmdName string, cmdArg ...string) (funcErrCode int, procErrCode int, e error) {
	//strArr := strings.Fields(strings.TrimSpace(cmdStr))
	//log.Println(strArr)
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
			logrus.Debug("stdout start")
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				stdOut <- scanner.Text()
			}
			outErr := stdout.Close()
			if outErr != nil {
				logrus.Debug("outerr := stdout.Close() err:", err)
			}
			logrus.Debug("stdout end")
		}()

		go func() {
			logrus.Debug("stderr start")
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				stdErr <- scanner.Text()
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
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			fmt.Println("exec.ExitError ExitCode: ", exitCode)
			return 0, exitCode, nil
		} else {
			fmt.Println("exec.ExitError: ", err)
			return 0, 0, errors.Wrap(err, "")
		}
	}

	log.Println("cmd.Wait end ok")
	return 0, 0, nil
}
