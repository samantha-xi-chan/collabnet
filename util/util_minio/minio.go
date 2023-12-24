package util_minio

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
)

type FileManager struct {
	endpoint        string
	accessKeyId     string
	secretAccessKey string
	useSSL          bool

	//bucketName string

	minioClient *minio.Client
	isConnected bool
	err         error
}

func CreateBucketIfNotExists(ctx context.Context, endpoint string, accessKeyID string, secretAccessKey string, useSSL bool, bucketName string) (e error) {
	minioClient, e := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if e != nil {
		log.Println("minio.New error: ", e)
		return e
	}

	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Println("BucketExists error: ", e)
		return err
	}
	if exists {
		log.Printf("bucket %s exists already", bucketName)
		return nil
	}
	e = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
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

func (f *FileManager) InitFM(ctx context.Context, endpoint string, accessKeyID string, secretAccessKey string, useSSL bool, bucketName string, clean bool) (e error) {
	// Initialize minio client object.
	log.Println("InitFM()")
	f.endpoint = endpoint
	f.accessKeyId = accessKeyID
	f.secretAccessKey = secretAccessKey
	f.useSSL = useSSL

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

	f.checkHealth(ctx)

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
	f.checkHealth(ctx)

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

func (f *FileManager) checkHealth(ctx context.Context) (e error) {
	if f.minioClient == nil {
		log.Println("ERROR: f.minioClient == nil") // todo: coding style
		return
	}

	if f.minioClient.IsOnline() == false {
		log.Println("ERROR: f.minioClient.IsOnline() == false") // todo: coding style

		f.minioClient, e = minio.New(f.endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(f.accessKeyId, f.secretAccessKey, ""),
			Secure: f.useSSL,
		})

		if e != nil {
			log.Println("f.minioClient, e = minio.New：e = ", e.Error())
			log.Fatal("Exit intentional ！！！")
		}
	}

	return
}

func (f *FileManager) DownloadFile(ctx context.Context, bucketName string, filePath string, objectName string) (e error) {
	f.checkHealth(ctx)

	err := f.minioClient.FGetObject(ctx, bucketName, objectName, filePath, minio.GetObjectOptions{})
	if err != nil {
		log.Println("FGetObject: ", err, ", objectName ", objectName)
		return err
	}

	log.Printf("Successfully FGetObject %s  \n", objectName)
	return nil
}
