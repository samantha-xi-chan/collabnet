package docker_vol

import (
	"collab-net-v2/package/util/util_minio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
)

func CreateVolumeFromFile(ctx context.Context, volumeName string, fileName string, fileContent string) (e error) {
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

	volume, err := cli.VolumeCreate(ctx, volume.CreateOptions{
		ClusterVolumeSpec: nil,
		Driver:            "",
		DriverOpts:        nil,
		Labels:            nil,
		Name:              volumeName,
	})
	if err != nil {
		log.Println("volume.Mountpoint: ", err)
		return err
	}

	log.Printf("Created volume: %s\n", volume.Name)

	volumeInspectResp, err := cli.VolumeInspect(context.Background(), volume.Name)
	if err != nil {
		log.Println("vol: ", volumeInspectResp)
		return err
	}

	log.Printf("VolumeMountpoint  Info: %+v\n", volumeInspectResp.Mountpoint)

	path := fmt.Sprintf("%s/%s", volume.Mountpoint, fileName)
	err = ioutil.WriteFile(path, []byte(fileContent), os.ModePerm)
	if err != nil {
		return err
	}

	//fmt.Printf("File '%s' written to volume '%s'\n", fileName, volume.Name)

	return nil
}

func CreateVolumeFromObjId(ctx context.Context, bucketName string, volumeName string, objId string, removeOldVol bool) (e error) {

	if objId == "" {
		log.Println("Warning: objId empty")
		return nil
	}

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

	if removeOldVol {
		err = cli.VolumeRemove(context.Background(), volumeName, true)
		if err != nil {
			return err
		}
		fmt.Printf("Volume %s deleted\n", volumeName)
	}

	volume, err := cli.VolumeCreate(ctx, volume.CreateOptions{
		ClusterVolumeSpec: nil,
		Driver:            "",
		DriverOpts:        nil,
		Labels:            nil,
		Name:              volumeName,
	})
	if err != nil {
		log.Println("volume.Mountpoint: ", err)
		return err
	}

	log.Printf("Created volume: %s\n", volume.Name)

	volumeInspectResp, err := cli.VolumeInspect(context.Background(), volume.Name)
	if err != nil {
		log.Println("vol: ", volumeInspectResp)
		return err
	}

	log.Printf("Volume Info: %+v\n", volumeInspectResp)

	util_minio.RestoreDir(bucketName, objId, volume.Mountpoint)

	return nil
}

func GetVolAbsPath(ctx context.Context, volumeName string) (AbsPath string, e error) {
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

	volumeInspectResp, err := cli.VolumeInspect(context.Background(), volumeName)
	if err != nil {
		log.Println("vol: ", volumeInspectResp)
		return "", err
	}

	log.Printf("Volume Info: %+v\n", volumeInspectResp)

	return volumeInspectResp.Mountpoint, nil
}

func RemoveVol(ctx context.Context, volumeId string) (e error) {
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

	if err := cli.VolumeRemove(context.Background(), volumeId, true); err != nil {
		log.Printf("无法删除卷: %v\n", err)
		return err
	}

	log.Printf("vol deleted volumeId= %s\n", volumeId)
	return nil
}

func IsVolExist(ctx context.Context, volumeName string) (exist bool, e error) {
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

	volumes, err := cli.VolumeList(context.Background(), volume.ListOptions{})
	if err != nil {
		log.Printf("Failed to list Docker volumes: %v", err) //
		return false, err
	}

	for _, volume := range volumes.Volumes {
		if volume.Name == volumeName {
			return true, nil
		}
	}

	return false, nil
}
