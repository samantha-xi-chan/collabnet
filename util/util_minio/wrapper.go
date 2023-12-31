package util_minio

import (
	"collab-net-v2/util/compress"
	"context"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
)

var fileManager FileManager

func CreateBucketIfNotExist(ctx context.Context, endpoint string, accessKeyID string, secretAccessKey string, bucketName string, objName string) (e error) {
	useSSL := false
	return CreateBucketIfNotExistsWithDefaultSign(ctx, endpoint, accessKeyID, secretAccessKey, useSSL, bucketName, objName)
}

func InitDistFileMs(ctx context.Context, endpoint string, accessKeyID string, secretAccessKey string, bucketName string, clean bool) (e error) {
	useSSL := false
	return fileManager.InitFM(ctx, endpoint, accessKeyID, secretAccessKey, useSSL, bucketName, clean)
}

func DeleteObjFromBucket(ctx context.Context, endpoint string, accessKeyID string, secretAccessKey string, bucketName string, objs []string) (e error) {
	useSSL := false
	return DeleteObjsFromBucket(ctx, endpoint, accessKeyID, secretAccessKey, useSSL, bucketName, objs)
}

func DeleteObjPrefixFromBucket(ctx context.Context, endpoint string, accessKeyID string, secretAccessKey string, bucketName string, objprefix string) (e error) {
	useSSL := false
	return DeleteObjPrefixsFromBucket(ctx, endpoint, accessKeyID, secretAccessKey, useSSL, bucketName, objprefix)
}

func IsConnected(ctx context.Context) (bool, error) {

	if fileManager.minioClient == nil {
		log.Println("ERROR: f.minioClient == nil") // todo: coding style
		return false, errors.New("fileManager.minioClient == nil: ")
	}

	return true, nil
}

func BackupDir(bucketName string, localDir string, objId string) error {
	ctx := context.Background()

	tmpFile, err := ioutil.TempFile("", "simple")
	if err != nil {
		return errors.Wrap(err, "ioutil.TempFile: ")
	}
	defer func() {
		tmpFile.Close()
		err := os.Remove(tmpFile.Name())
		if err != nil {
			log.Println("Error cleaning up temporary file:", err)
		}
	}()

	if err := compress.TarFiles(localDir, tmpFile.Name()); err != nil {
		return errors.Wrap(err, "util_zip.TarFileOrDir: ")
	}

	log.Println("BackupDir", bucketName, tmpFile.Name(), objId)
	if err := fileManager.UploadFile(ctx, bucketName, tmpFile.Name(), objId); err != nil {
		return errors.Wrap(err, "fileManager.UploadFile: ")
	}

	return nil
}

func RestoreDir(bucketName string, objId string, localDir string) error {
	log.Println("RestoreDir input: objId =", objId, ", localDir:", localDir)
	defer log.Println("RestoreDir end: localDir =", localDir, ", objId:", objId)

	ctx := context.Background()

	tmpFile, err := ioutil.TempFile("", "simple")
	if err != nil {
		log.Println("ERROR: tmpFile, err := ioutil.TempFile, err =", err.Error())
		return err
	}
	defer func() {
		tmpFile.Close()
		err := os.Remove(tmpFile.Name())
		if err != nil {
			log.Println("Error cleaning up temporary file:", err)
		}
	}()

	if err := fileManager.DownloadFile(ctx, bucketName, tmpFile.Name(), objId); err != nil {
		return err
	}

	compress.UntarFiles(tmpFile.Name(), localDir)

	return nil
}
