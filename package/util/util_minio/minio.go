package util_minio

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
)

type FileManager struct {
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	useSSL          bool

	//bucketName string

	minioClient *minio.Client
	isConnected bool
	err         error
}

func (f *FileManager) InitFM(ctx context.Context, endpoint string, accessKeyID string, secretAccessKey string, useSSL bool, bucketName string, clean bool) (e error) {
	// Initialize minio client object.
	f.minioClient, e = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if e != nil {
		log.Println("minio.New error: ", e)
		return e
	}

	if clean {
		log.Println("InitFM clean ing ")
		objectsCh := make(chan string)
		go func() {
			defer close(objectsCh)
			for object := range f.minioClient.ListObjects(context.TODO(), bucketName, minio.ListObjectsOptions{Recursive: true}) {
				if object.Err != nil {
					log.Println(object.Err)
					return
				}
				objectsCh <- object.Key
			}
		}()

		for objectKey := range objectsCh {
			err := f.minioClient.RemoveObject(context.TODO(), bucketName, objectKey, minio.RemoveObjectOptions{})
			if err != nil {
				log.Printf("Error deleting object %s: %v\n", objectKey, err)
			}
		}

		if e := f.minioClient.RemoveBucket(ctx, bucketName); e != nil {
			log.Println("RemoveBucket e: ", e)
		}
	}

	exists, err := f.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Println("BucketExists error: ", e)
		return err
	}
	if exists {
		log.Printf("bucket %s exists already", bucketName)
		return nil
	}
	e = f.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region:        "",
		ObjectLocking: false,
	})
	if e != nil {
		log.Println("minio.MakeBucket error: ", e)
		return e
	}

	log.Printf("bucket %s created", bucketName)
	return nil
}

func (f *FileManager) RemoveBucket(ctx context.Context, bucketName string) (e error) {
	err := f.minioClient.RemoveBucket(ctx, bucketName)
	if err != nil {
		log.Println("RemoveBucket: ", err)
		return err
	}

	log.Printf("bucket %s Removed", bucketName)
	return nil
}

// local dir : remote url
func (f *FileManager) UploadFile(ctx context.Context, bucketName string, filePath string, objectName string) (e error) {
	contentType := "application/zip"

	// Upload the zip file with FPutObject
	_, err := f.minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Println("FPutObject err: ", err)
		return err
	}

	//log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)

	return nil
}

func (f *FileManager) DownloadFile(ctx context.Context, bucketName string, filePath string, objectName string) (e error) {

	// Upload the zip file with FPutObject
	err := f.minioClient.FGetObject(ctx, bucketName, objectName, filePath, minio.GetObjectOptions{})
	if err != nil {
		log.Println("FGetObject: ", err, ", objectName ", objectName)
		return err
	}

	log.Printf("Successfully FGetObject %s  \n", objectName)
	return nil
}
