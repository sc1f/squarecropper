package main

import (
	"os",
	"encoding/json",
	"errors",
	"fmt",
	"context",
	"net/http",
	"image",
	"image/jpeg", // spotify only supports JPEGs for cropping
	"github.com/disintegration/imaging",
	"github.com/museli/smartcrop",
	"github.com/muesli/smartcrop/nfnt",
	"github.com/aws/aws-lambda-go",
	"github.com/aws/aws-lambda-go/events",
)

type Response struct {
	CroppedImageUrl string `json:"cropped_image_url"`,
}

type Image struct {
	Url string

}

/* 
bucket = record['s3']['bucket']['name']
        key = record['s3']['object']['key'] 
        download_path = '/tmp/{}{}'.format(uuid.uuid4(), key)
        upload_path = '/tmp/resized-{}'.format(key)
        
        s3_client.download_file(bucket, key, download_path)
        resize_image(download_path, upload_path)
        s3_client.upload_file(upload_path, '{}resized'.format(bucket), key)
*/
        
func downloadImage(image_url string) Image {

}

func cropImage(image_to_crop Image) Image {

}

func uploadImageToS3(image_to_save Image) bool {

}

func handler(ctx context, s3Event events.S3Event) (Response, err) {
	for _, record := range s3Event.Records {

	}
	/* 
	for _, record := range s3Event.Records {
        s3 := record.S3
        fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3.Bucket.Name, s3.Object.Key) 
    }
	*/
}

func main() {
	lambda.Start(handler)
}