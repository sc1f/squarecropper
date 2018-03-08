package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
	"github.com/twinj/uuid"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
)

type Response struct {
	Success         bool   `json:"success"`
	CroppedImageUrl string `json:"cropped_image_url"`
}

// type SubImager is used by smartcrop to hold the cropped image
type SubImager interface {
	// SubImage returns an image object from SmartCrop
	SubImage(r image.Rectangle) image.Image
}

type Image struct {
	err error
	// S3 Bucket
	Bucket        string `json:"bucket"`
	CroppedBucket string `json:"cropped_bucket"`
	// S3 Key
	Key        string `json:"key"`
	CroppedKey string `json:"cropped_key"`
	// Filesystem information
	FileName    string `json:"filename"`
	Path        string `json:"path"`
	CroppedPath string `json:"cropped_path"`
}

func (img *Image) downloadImage() {
	// Abstract error checking
	if img.err != nil {
		return
	}
	// Create a file to write the S3 Object contents to
	img_uuid = string(uuid.NewV4())
	file_name = fmt.Sprintf("%s%s", img_uuid, img.Key)
	file_path = "/tmp/" + file_name

	file, err := os.Create(file_path)
	if err != nil {
		img.err = fmt.Errorf("Failed to create file %q: %v", file_path, err)
		return
	}

	// Add the name and path to our final image object
	img.Filename = file_name
	img.Path = file_path

	// Write the contents of S3 Object to the file
	bytes, err := downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(img.Bucket),
		Key:    aws.String(img.Key),
	})

	if err != nil {
		img.err = fmt.Errorf("Failed to download image: %v", err)
		return
	}

	fmt.Printf("downloaded image, %d bytes\n", bytes)
}

func (img *Image) cropImage() err {
	if img.err != nil {
		return
	}

	file, _ = os.Open(&img.Path)
	defer file.Close()

	decoded_img, _, err = image.Decode(file)
	if err != nil {
		img.err = err
		return
	}

	crop_analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	best_crop, _ := analyzer.FindBestCrop(decoded_img, 500, 500)

	cropped_img := cropped_img.(SubImager).SubImage(best_crop)
	cropped_img_path = "/tmp/resized-" + img.FileName

	cropped_img_file, err := os.Create(cropped_img_path)
	defer cropped_img_file.Close()

	if err != nil {
		img.err = fmt.Errorf("failed to create file %q, %v", cropped_img_path, err)
		return
	}

	jpeg.Encode(cropped_img_file, cropped_img, &jpeg.Options{Quality: 100})
	img.CroppedPath = cropped_img_path
}

func (img *Image) uploadCroppedImageToS3() (string, err) {
	if img.err != nil {
		return
	}

	img_file, err := os.Open(img.CroppedPath)
	if err != nil {
		img.err = fmt.Errorf("failed to open file %q, %v", filename, err)
		return
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(img.ResizedBucket),
		Key:    aws.String(img.CroppedKey),
		Body:   img_file,
	})

	if err != nil {
		img.err = fmt.Errorf("failed to upload file, %v", err)
		return
	}

	fmt.Printf("file uploaded to, %s\n", aws.StringValue(result.Location))
	return img.CroppedKey, nil
}

func handler(ctx context, s3Event events.S3Event) (Response, err) {
	// Initialize a new S3 session + downloader for the image
	session, _ := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(session)
	uploader := s3manager.NewUploader(session)

	// Grab from S3
	record := s3Event.Records[0]
	s3 := record.s3
	bucket := s3.Bucket.Name
	resized_bucket := "resized-" + s3.Bucket.Name
	key := s3.Bucket.Key

	img, err := Image{bucket, key, nil, nil}
	img.downloadImage()
	img.cropImage()
	img.uploadCroppedImageToS3()
	if img.err != nil {
		return nil, img.err
	}

	return Response{img.CroppedKey}, nil
}

func main() {
	lambda.Start(handler)
}
