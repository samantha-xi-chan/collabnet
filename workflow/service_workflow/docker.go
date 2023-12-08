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

func CreateContainerWrapper(ctx context.Context, req api.PostContainerReq, containerName string) (containerId string, err error) {
	log.Println("PostContainer req: ", req)

	taskId := req.TaskId

	arrayB := make([]string, len(req.CmdStr)+2)

	arrayB[0] = "/bin/sh"
	arrayB[1] = "/path/in/docker/cmd.sh"
	copy(arrayB[2:], req.CmdStr[0:])

	// check if image ready
	exists, e := docker_image.IsImageExists(ctx, req.Image)
	if e != nil {
		return "", errors.Wrap(e, "docker_image.IsImageExists")
	}

	if exists {
		log.Println("docker_image.IsImageExists true", req.Image) // just debug info
	} else {
		log.Println("docker_image.PullImageBlo : ", req.Image) //  INFO
		ee := docker_image.PullImageBlo(ctx, req.Image)        // todo ： 添加 提前下载 image的能力到产品接口
		if ee != nil {
			return "", errors.Wrap(ee, "docker_image.PullImageBlo: ")
		}
	}

	for idx, val := range req.BindIn {
		log.Printf("for idx, val := range req.BindIn idx= %d  containerName: %s \n", idx, containerName)
		if val.VolId != "" {
			volName := fmt.Sprintf("%s_%s", containerName, val.VolId)
			ee := docker_vol.CreateVolumeFromObjId(ctx, req.BucketName, volName, val.VolId, false)
			if ee != nil {
				log.Println("CreateVolumeFromObjId e=", ee)
				return "", errors.Wrap(ee, "docker_vol.CreateVolumeFromObjId")
			}
		}
	}

	numCPU := runtime.NumCPU()
	cpuSet := "0"
	if numCPU/2-1 >= 1 {
		cpuSet = fmt.Sprintf("0-%d", numCPU/2-1)
	}

	id, err := docker_container.CreateContainer(ctx,
		taskId, req.CbAddr,
		false,
		req.Image,
		arrayB,
		8*1000, // todo： 改为全局可配置
		200,    // todo： 改为全局可配置
		cpuSet,
		containerName,

		req.BindIn,
		req.BindOut,
	)
	if err != nil {
		return "", err
	}

	containerId = id[:12]
	log.Println("docker_container.CreateContainer, containerId = ", containerId)

	return containerId, nil
}

func StartContainerAndWait(ctx context.Context, containerId string, req api.PostContainerReq, containerName string) (exitCode int, e error) {
	logRt := req.LogRt
	cleanContainer := req.CleanContainer

	e = docker_container.StartContainer(ctx, containerId)
	if e != nil {
		return 0, errors.Wrap(e, "docker_container.StartContainerVV: ")
	}

	exitCode, e = docker_container.WatchContainer(ctx, req.TaskId, containerId, cleanContainer, logRt)
	if e != nil {
		log.Println(" !!! docker_container.WatchContainer, err = ", e)
		return 0, e
	}
	log.Println("exitCode: ", exitCode, " , containerId: ", containerId)

	for idx, val := range req.BindOut {
		log.Printf("BindOut, idx: %d, val:%s", idx, val)

		volName := fmt.Sprintf("%s_%s", containerName, val.VolId)
		objId := val.VolId

		absPath, e := docker_vol.GetVolAbsPath(ctx, volName)
		if e != nil {
			log.Println("docker_vol.GetVolAbsPath, err = ", e)
		}
		e = util_minio.BackupDir(req.BucketName, absPath+"/", objId)
		if e != nil {
			log.Println("docker_image.BackupDir, err = ", e)
		}
	}

	if cleanContainer {
		for idx, val := range req.BindIn {
			log.Printf("deleting BindIn, idx: %d, val:%s", idx, val)
			newVolId := fmt.Sprintf("%s_%s", containerName, val.VolId)
			docker_vol.RemoveVol(ctx, newVolId)
		}
		for idx, val := range req.BindOut {
			log.Printf("deleting BindOut, idx: %d, val:%s", idx, val)
			newVolId := fmt.Sprintf("%s_%s", containerName, val.VolId)
			docker_vol.RemoveVol(ctx, newVolId)
		}
	}

	return exitCode, nil
}
