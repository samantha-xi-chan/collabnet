package service_workflow

import (
	"collab-net-v2/api"
	"collab-net-v2/package/util/docker_container"
	"collab-net-v2/package/util/docker_image"
	"collab-net-v2/package/util/docker_vol"
	"collab-net-v2/package/util/util_minio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"runtime"
)

func PostContainer(ctx context.Context, req api.PostContainerReq) (containerId string, err error) {
	log.Println("PostContainer req: ", req)

	taskId := req.TaskId
	logRt := req.LogRt
	cleanContainer := req.CleanContainer

	arrayB := make([]string, len(req.CmdStr)+2)

	arrayB[0] = "/bin/sh"
	arrayB[1] = "/path/in/docker/cmd.sh"
	//arrayB[2] = "enable_debug"
	copy(arrayB[2:], req.CmdStr[0:])

	// check if image ready
	exists, e := docker_image.IsImageExists(ctx, req.Image)
	if e != nil {
		return "", errors.Wrap(e, "docker_image.IsImageExists")
	}

	if exists {
		log.Println("docker_image.IsImageExists true") // just debug info
	} else {
		log.Println("docker_image.PullImageBlo : ", req.Image) //  INFO
		ee := docker_image.PullImageBlo(ctx, req.Image)        // todo ： 添加 提前下载 image的能力到产品接口
		if ee != nil {
			return "", errors.Wrap(ee, "docker_image.PullImageBlo: ")
		}
	}

	//name := idgen.GetIDWithPref(config.CONTINAER_PREF)
	for idx, val := range req.BindIn {
		log.Println("for idx, val := range req.BindIn idx=", idx)
		if val.VolId != "" {
			ee := docker_vol.CreateVolumeFromObjId(ctx, req.BucketName, val.VolId, val.VolId, false)
			if ee != nil {
				log.Println("CreateVolumeFromObjId e=", ee)
				return "", errors.Wrap(ee, "docker_vol.CreateVolumeFromObjId")
			}
		}
	}

	numCPU := runtime.NumCPU()
	fmt.Printf("CPU core cnt：%d\n", numCPU)
	cpuSet := "0"
	if numCPU/2-1 >= 1 {
		cpuSet = fmt.Sprintf("0-%d", numCPU/2-1)
	}

	containerId, err = docker_container.StartContainer(ctx,
		taskId, req.CbAddr,
		false,
		req.Image,
		arrayB,
		800,
		50,
		cpuSet,
		req.Name, /* container name*/

		req.BindIn,
		req.BindOut,
	)
	if err != nil {
		return "", err
	}

	go func() {
		// v2 scope: NotifyTaskStatus(taskId, api.TASK_STATUS_RUNNING, api.EXIT_CODE_NONE)
		exitCode, e := docker_container.WatchContainer(ctx, req.TaskId, containerId, cleanContainer, logRt)
		if e != nil {
			log.Println(" !!! docker_container.WatchContainer, err = ", err)
		}
		log.Println("exitCode: ", exitCode)

		out := []string{}
		if len(req.BindOut) >= 1 {
			// get container expId path
			absPath, e := docker_vol.GetVolAbsPath(ctx, req.BindOut[0].VolId)
			if e != nil {
				log.Println("docker_vol.GetVolAbsPath, err = ", err)
			}
			err = util_minio.BackupDir(req.BucketName, absPath+"/", req.BindOut[0].VolId)
			if err != nil {
				log.Println("docker_image.BackupDir, err = ", err)
			}
			out = append(out, req.BindOut[0].VolId)
		}

		log.Println("out: ", out)

		// v2 scope:  NotifyTaskStatus(taskId, api.TASK_STATUS_END, exitCode)
	}()

	log.Printf("containerId=%s,  exitCode not ready", containerId)

	return containerId, nil
}

func PostContainerBlock(ctx context.Context, req api.PostContainerReq) (containerId string, exitCode int, err error) {
	log.Println("PostContainer req: ", req)

	taskId := req.TaskId
	logRt := req.LogRt
	cleanContainer := req.CleanContainer

	arrayB := make([]string, len(req.CmdStr)+2)

	arrayB[0] = "/bin/sh"
	arrayB[1] = "/path/in/docker/cmd.sh"
	//arrayB[2] = "enable_debug"
	copy(arrayB[2:], req.CmdStr[0:])

	// check if image ready
	exists, e := docker_image.IsImageExists(ctx, req.Image)
	if e != nil {
		return "", 0, errors.Wrap(e, "docker_image.IsImageExists")
	}

	if exists {
		log.Println("docker_image.IsImageExists true") // just debug info
	} else {
		log.Println("docker_image.PullImageBlo : ", req.Image) //  INFO
		ee := docker_image.PullImageBlo(ctx, req.Image)        // todo ： 添加 提前下载 image的能力到产品接口
		if ee != nil {
			return "", 0, errors.Wrap(ee, "docker_image.PullImageBlo: ")
		}
	}

	//name := idgen.GetIDWithPref(config.CONTINAER_PREF)
	for idx, val := range req.BindIn {
		log.Println("for idx, val := range req.BindIn idx=", idx)
		if val.VolId != "" {
			ee := docker_vol.CreateVolumeFromObjId(ctx, req.BucketName, val.VolId, val.VolId, false)
			if ee != nil {
				log.Println("CreateVolumeFromObjId e=", ee)
				return "", 0, errors.Wrap(ee, "docker_vol.CreateVolumeFromObjId")
			}
		}
	}

	numCPU := runtime.NumCPU()
	fmt.Printf("CPU core cnt：%d\n", numCPU)
	cpuSet := "0"
	if numCPU/2-1 >= 1 {
		cpuSet = fmt.Sprintf("0-%d", numCPU/2-1)
	}

	containerId, err = docker_container.StartContainer(ctx,
		taskId, req.CbAddr,
		false,
		req.Image,
		arrayB,
		800,
		50,
		cpuSet,
		req.Name, /* container name*/

		req.BindIn,
		req.BindOut,
	)
	if err != nil {
		return "", 0, err
	}

	// v2 scope: NotifyTaskStatus(taskId, api.TASK_STATUS_RUNNING, api.EXIT_CODE_NONE)
	exitCode, e = docker_container.WatchContainer(ctx, req.TaskId, containerId, cleanContainer, logRt)
	if e != nil {
		log.Println(" !!! docker_container.WatchContainer, err = ", err)
		return containerId, 0, err
	}
	log.Println("exitCode: ", exitCode)

	out := []string{}
	if len(req.BindOut) >= 1 {
		// get container expId path
		absPath, e := docker_vol.GetVolAbsPath(ctx, req.BindOut[0].VolId)
		if e != nil {
			log.Println("docker_vol.GetVolAbsPath, err = ", err)
		}
		err = util_minio.BackupDir(req.BucketName, absPath+"/", req.BindOut[0].VolId)
		if err != nil {
			log.Println("docker_image.BackupDir, err = ", err)
		}
		out = append(out, req.BindOut[0].VolId)
	}

	log.Println("out: ", out)

	// v2 scope:  NotifyTaskStatus(taskId, api.TASK_STATUS_END, exitCode)

	//log.Printf("containerId=%s,  exitCode not ready", containerId)

	return containerId, exitCode, nil
}
