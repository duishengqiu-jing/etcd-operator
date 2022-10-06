package file

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	ctrl "sigs.k8s.io/controller-runtime"
)

type s3Uploader struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
}

func NewS3Uploader(Endpoint, AK, SK string) *s3Uploader {
	return &s3Uploader{
		Endpoint:        Endpoint,
		AccessKeyID:     AK,
		SecretAccessKey: SK,
	}
}

// 初使化 minio client 对象
func (su *s3Uploader) InitClient() (*minio.Client, error) {
	return minio.New(su.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(su.AccessKeyID, su.SecretAccessKey, ""),
		Secure: true,
	})
}

func (su *s3Uploader) Upload(ctx context.Context, filePath string) (int64, error) {
	client, err := su.InitClient()
	if err != nil {
		return 0, err
	}
	log := ctrl.Log.WithName("backup-upload")
	bucketName := "duishengqiu"      // todo
	objectName := "etcd-snapshot.db" // todo
	location := "us-east-1"
	//log.Printf("--------s3Upload-----bucketName:%s,objectName:%s,")
	log.Info("[s3Upload] Uploading snapshot")
	err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Info("We already own")
		} else {
			log.Info("-----创建失败，bucket不存在-----")
		}
	} else {
		log.Info("Successfully created")
	}

	uploadInfo, err := client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return 0, err
	}
	return uploadInfo.Size, nil
}
