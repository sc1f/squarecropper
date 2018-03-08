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
	CroppedImageUrl string `json:"cropped_image_url"`
}

type Image struct {
	Bucket        string `json:"bucket"`
	CroppedBucket string `json:"cropped_bucket"`
	Key           string `json:"key"`
	CroppedKey    string `json:"cropped_key"`
	FileName      string `json:"filename"`
	Path          string `json:"path"`
	CroppedPath   string `json:"cropped_path"`
}

// type SubImager is used by smartcrop to hold the cropped image
type SubImager interface {
	// SubImage returns an image object from SmartCrop
	SubImage(r image.Rectangle) image.Image
}

func downloadImage(bucket string, key string) (Image, error) {
	image = Image{bucket, key, nil, nil}

	// Create a file to write the S3 Object contents to
	image_uuid = string(uuid.NewV4())
	file_name = fmt.Sprintf("%s%s", image_uuid, image.Key)
	file_path = "/tmp/" + file_name

	file, err := os.Create(file_path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %q, %v", file_path, err)
	}

	// Add the name and path to our final image object
	image.Filename = file_name
	image.Path = file_path

	// Write the contents of S3 Object to the file
	bytes, err := downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(image.Bucket),
		Key:    aws.String(image.Key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download image, %v", err)
	}

	fmt.Printf("downloaded image, %d bytes\n", bytes)

	return image, nil
}

func cropImage(image Image) (Image, err) {
	file, _ = os.Open(image.Path)
	defer file.Close()

	decoded_image, _, err = image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf(err)
	}

	crop_analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	best_crop, _ := analyzer.FindBestCrop(decoded_image, 500, 500)

	cropped_image := decoded_image.(SubImager).SubImage(best_crop)

	cropped_image_filename = "resized-" + image.FileName
	cropped_image_filepath = "/tmp/" + cropped_image_filename

	cropped_image_file, err := os.Create(cropped_image_filepath)
	defer cropped_image_file.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to create file %q, %v", cropped_image_filepath, err)
	}

	jpeg.Encode(cropped_image_file, cropped_image, &jpeg.Options{Quality: 100})
	image.CroppedPath = cropped_image_filepath
	return image
}

func uploadCroppedImageToS3(image Image) (string, err) {
	image_file, err := os.Open(image.CroppedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q, %v", filename, err)
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(image.ResizedBucket),
		Key:    aws.String(image.CroppedKey),
		Body:   image_file,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload file, %v", err)
	}

	fmt.Printf("file uploaded to, %s\n", aws.StringValue(result.Location))
	return image.CroppedKey, nil
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

	// pass metadata to the image download/crop/upload workflow
	downloaded_image, err := downloadImage(bucket, key)
	if err != nil {
		return nil, err
	}

	cropped_image, err := cropImage(downloaded_image)
	if err != nil {
		return nil, err
	}

	upload_success, err := uploadCroppedImageToS3(cropped_image)
	if err != nil {
		return nil, err
	}

	response := Response{cropped_image.CroppedKey}

	return response, nil
}

func main() {
	lambda.Start(handler)
}
