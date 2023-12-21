package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// AWS S3 configuration
	awsRegion := "ap-south-1"
	awsBucket := "files-input-v1"
	awsAccessKey := "AKIAUW7IC24WI7EUSU5T"
	awsSecretKey := "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"

	s3Client := initS3Client(awsRegion, awsAccessKey, awsSecretKey)

	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/", func(c *gin.Context) {

		fileName := c.PostForm("fileName")
		textInput := c.PostForm("textInput")
		if fileName == "" {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Please enter a file name",
			})
			return
		}

		// Save the text input to a file with the provided name
		err := ioutil.WriteFile("./assets/uploads/"+fileName+".txt", []byte(textInput), 0644)
		if err != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to save text input",
			})
			return
		}
		// Upload the file to S3
		err = uploadFileToS3(s3Client, fileName+".txt", "./assets/uploads/"+fileName+".txt", awsBucket)
		if err != nil {
			fmt.Println("Error uploading file to S3:", err)

		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"text":     textInput,
			"fileName": fileName,
		})
	})
	r.Run() // listen and serve 80
}

func initS3Client(region, accessKey, secretKey string) *s3.S3 {
	sess, err := session.NewSession(&aws.Config{ //
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		fmt.Println("Error creating AWS session:", err)
		os.Exit(1)
	}

	svc := s3.New(sess)
	return svc
}

func uploadFileToS3(svc *s3.S3, fileName, filePath, bucketName string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println("Bucket does not exist")
			case s3.ErrCodeBucketAlreadyExists:
				fmt.Println("Bucket already exists")
			default:
				fmt.Println("Error uploading file to S3:", aerr.Error())
			}
		} else {
			fmt.Println("Error uploading file to S3:", err.Error())
		}
		return err
	}
	fmt.Println("File uploaded successfully to S3")
	return nil
}
