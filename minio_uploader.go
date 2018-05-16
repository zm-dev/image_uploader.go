package image_uploader

import (
	"github.com/minio/minio-go"
	"mime"
)

type minioUploader struct {
	h           Hasher
	s           Store
	minioClient *minio.Client
	bucketName  string
}

func (mu *minioUploader) saveToMinio(hashValue string, fh FileHeader, info ImageInfo) error {
	// 在 apline 镜像中 mime.TypeByExtension 只能用 jpg
	if info.format == "jpeg" {
		info.format = "jpg"
	}
	_, err := mu.minioClient.PutObject(
		mu.bucketName,
		hashValue,
		fh.File,
		fh.Size,
		minio.PutObjectOptions{ContentType: mime.TypeByExtension("." + info.format)},
	)
	return err
}

func (mu *minioUploader) Upload(fh FileHeader) (*Image, error) {
	info, err := DecodeImageInfo(fh.File)
	if err != nil {
		return nil, err
	}

	hashValue, err := mu.h.Hash(fh.File)
	if err != nil {
		return nil, err
	}
	if exist, err := mu.s.ImageExist(hashValue); exist && err == nil {
		// 图片已经存在
		return mu.s.ImageLoad(hashValue)
	} else if err != nil {
		return nil, err
	}

	if err := mu.saveToMinio(hashValue, fh, info); err != nil {
		return nil, err
	}

	return saveToStore(mu.s, hashValue, fh.Filename, info)
}

func NewMinioUploader(h Hasher, s Store, minioClient *minio.Client, bucketName string) Uploader {
	return &minioUploader{h: h, s: s, minioClient: minioClient, bucketName: bucketName}
}