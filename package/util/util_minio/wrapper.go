package util_minio

import (
	"collab-net-v2/package/util/util_zip"
	"context"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
)

var fileManager FileManager

func Init(ctx context.Context, endpoint string, accessKeyID string, secretAccessKey string, bucketName string, clean bool) (e error) {
	useSSL := false
	return fileManager.InitFM(ctx, endpoint, accessKeyID, secretAccessKey, useSSL, bucketName, clean)
}

// 备份文件夹内部
func BackupDir(bucketName string, localDir string, objId string) (x error) {
	//log.Println("BackupDir input: localDir = ", localDir, ", objId: ", objId)
	//defer log.Println("BackupDir end: localDir = ", localDir, ", objId: ", objId)
	ctx := context.Background()

	tmpFile, err := ioutil.TempFile("", "simple")
	if err != nil {
		return errors.Wrap(err, "ioutil.TempFile: ")
	}
	defer tmpFile.Close()

	e := util_zip.RecursiveZip(localDir, tmpFile.Name())
	if e != nil {
		return errors.Wrap(e, "util_zip.RecursiveZip: ")
	}

	log.Println("BackupDir", bucketName, tmpFile.Name(), objId)
	e = fileManager.UploadFile(ctx, bucketName, tmpFile.Name(), objId)
	if e != nil {
		return errors.Wrap(e, "fileManager.UploadFile: ")
	}
	return nil
}

func RestoreDir(bucketName string, objId string, localDir string) (e error) {
	log.Println("RestoreDir input: objId = ", objId, ", localDir: ", localDir)
	defer log.Println("RestoreDir end: localDir = ", localDir, ", objId: ", objId)
	ctx := context.Background()

	tmpFile, _ := ioutil.TempFile("", "simple")
	defer tmpFile.Close()
	fileManager.DownloadFile(ctx, bucketName, tmpFile.Name(), objId)

	util_zip.RecursiveUnzip(tmpFile.Name(), localDir)

	return nil
}
