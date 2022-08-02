package main

import (
	"encoding/json"
	"github.com/aldelo/common/wrapper/aws/awsregion"
	"github.com/aldelo/common/wrapper/s3"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	r := gin.Default()
	r.GET("/uploadTiktok", func(c *gin.Context) {
		urlPath := c.Query("q")

		// yt-dlp --referer "https://www.tiktok.com" "https://www.tiktok.com/t/ZTRUJ1NmF/?k=1" --extract-audio --audio-format mp3 -o "asd.mp3"
		if len(urlPath) == 0 {
			c.JSON(500, "invalid URL")
		}
		log.Println("Got request for ", urlPath)

		err := exec.Command("yt-dlp", "--referer", "\"https://www.tiktok.com\"", urlPath, "--extract-audio", "--audio-format", "mp3", "-o", "%(title)s.%(ext)s").Run()
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error processing URL : " + urlPath,
			})

			return
		}else{
			file := WalkMatch()
			log.Println("THE TITLE IS: " + file)

			urlPath := strings.Replace(urlPath, "https://www.tiktok.com/t/", "", -1)
			urlPath = strings.Replace(urlPath, "/?k=1", "", -1)

			s := s3.S3{
				AwsRegion:   awsregion.AWS_us_west_2_oregon,
				HttpOptions: nil,
				BucketName:  "calc.masa.space",
			}
			err = s.Connect()
			if err != nil {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})

				return
			}
			f, err := os.Open(file)
			if err != nil {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})

				return
			}
			defer f.Close()

			byteContainer, err := ioutil.ReadAll(f)
			if err != nil {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})

				return
			}

			location, err := s.Upload(nil, byteContainer, "music/" + urlPath + ".mp3")

			if err != nil {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})

				return
			}

			log.Println(location)

			e := os.Remove(file)
			if e != nil {
				c.JSON(500, gin.H{
					"message": e.Error(),
				})

				return
			}


			c.JSON(200, file)

		}
	})
	r.GET("/upload", func(c *gin.Context) {

		urlPath := c.Query("q")

		if len(urlPath) == 11 && urlPath != "favicon.ico" {
			log.Printf("Got request for : %s", urlPath)

			err, valid := IsVideoValid(urlPath)
			if err != nil {
				c.JSON(500, err.Error())

				return
			}

			if !valid {
				c.JSON(500, "Video too long")

				return
			}

			err = exec.Command("yt-dlp", "--extract-audio", "--audio-format", "mp3", "--output", "%(title)s.%(ext)s", urlPath).Run()
			if err != nil {
				c.JSON(500, gin.H{
					"message": "Error processing URL : " + urlPath,
				})

				return
			}else{

				file := WalkMatch()
				log.Println("THE TITLE IS: " + file)

				s := s3.S3{
					AwsRegion:   awsregion.AWS_us_west_2_oregon,
					HttpOptions: nil,
					BucketName:  "calc.masa.space",
				}
				err = s.Connect()
				if err != nil {
					c.JSON(500, gin.H{
						"message": err.Error(),
					})

					return
				}
				f, err := os.Open(file)
				if err != nil {
					c.JSON(500, gin.H{
						"message": err.Error(),
					})

					return
				}
				defer f.Close()

				byteContainer, err := ioutil.ReadAll(f)
				if err != nil {
					c.JSON(500, gin.H{
						"message": err.Error(),
					})

					return
				}

				location, err := s.Upload(nil, byteContainer, "music/" + urlPath+".mp3")

				if err != nil {
					c.JSON(500, gin.H{
						"message": err.Error(),
					})

					return
				}

				log.Println(location)

				e := os.Remove(file)
				if e != nil {
					c.JSON(500, gin.H{
						"message": e.Error(),
					})

					return
				}

				c.JSON(200, strings.Replace(file, "\"", "", -1))

			}

		} else{
			c.JSON(500, gin.H{
				"message": "Bad URL : " + urlPath,
			})

			return
		}

	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, "OK")
	})

	log.Printf("Youtube MP3 Download Backend Server Started 2")
	r.Run() // listen and serve on 0.0.0.0:8080
}


func WalkMatch() string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if strings.Contains(f.Name(), ".mp3") {
			return f.Name()
		}
	}

	return ""
}

func IsVideoValid(url string) (error, bool) {
	resp, err := http.Get("https://www.googleapis.com/youtube/v3/videos?id=" + url + "&part=contentDetails&key=AIzaSyAfAI5KU0Cmh6oKOeAjXskv4yMfc4Xzg8k")
	if err != nil {
		return err, false
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, false
	}

	var yt YoutubeData
	//Convert the body to type string
	err = json.Unmarshal(body, &yt)
	if err != nil {
		return err, false
	}

	h := yt.Items[0].ContentDetails.Duration
	if strings.Contains(h, "H") {
		return nil, false
	}
	return nil, true
}


type YoutubeData struct {
	Items []YoutubeChild `json:"items"`
}

type YoutubeChild struct {
	ContentDetails ContentDetails `json:"contentDetails"`
}

type ContentDetails struct {
	Duration string `json:"duration"`
}

//  GOOS=linux GOARCH=amd64 go build -o lysten_api_linux
// ssh -i ~/.ssh/lysten.pem ubuntu@50.18.240.5
// scp -i ~/.ssh/lysten.pem lysten_api_linux ubuntu@50.18.240.5:/home/ubuntu/api
// systemctl --lines=5000 status lystenapi
// 50.18.240.5