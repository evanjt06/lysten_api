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

//func healthHandler(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprintf(w, "OK")
//}
//
//func handler(w http.ResponseWriter, r *http.Request) {
//
//	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
//	w.Header().Set("Pragma", "no-cache")
//	w.Header().Set("Expires", "0")
//
//	urlPath := r.FormValue("q")
//
//	if len(urlPath) == 11 && urlPath != "favicon.ico" {
//		log.Printf("Got request for : %s", urlPath)
//		err := exec.Command("yt-dlp", "--extract-audio", "--audio-format", "mp3", "--output", "%(title)s.%(ext)s", urlPath).Run()
//		if err != nil {
//			log.Printf("Error occurred processing URL : %s", urlPath)
//		}else{
//
//			file := WalkMatch()
//			log.Println("THE TITLE IS: " + file)
//
//			s := s3.S3{
//				AwsRegion:   awsregion.AWS_us_west_2_oregon,
//				HttpOptions: nil,
//				BucketName:  "calc.masa.space",
//			}
//			err = s.Connect()
//			if err != nil {
//				log.Fatal("could not connect to s3 - aws")
//			}
//			f, err := os.Open(file)
//			if err != nil {
//				log.Fatal(err)
//			}
//			defer f.Close()
//
//			byteContainer, err := ioutil.ReadAll(f)
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			location, err := s.Upload(nil, byteContainer, "music/" + urlPath+".mp3")
//
//
//
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			log.Println(location)
//
//			w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
//			w.Header().Set("Content-Type", "audio/mp3")
//			//http.ServeFile(w, r, r.URL.Path[1:]+".mp3")
//			// return file (or the title of the video)
//			_, _ = fmt.Fprintf(w, file)
//
//			e := os.Remove(file)
//			if e != nil {
//				log.Fatal(e)
//			}
//		}
//
//	} else{
//		log.Printf("Bad URL : %s", urlPath)
//	}
//}
//func main() {
//	mux := http.NewServeMux()
//
//	mux.HandleFunc("/upload", handler)
//	mux.HandleFunc("/health", healthHandler)
//
//	log.Printf("Youtube MP3 Download Backend Server Started 2")
//
//	err := http.ListenAndServe(":8080", mux)
//
//	if err != nil {
//		log.Fatal(err.Error())
//	}
//}
//
//func WalkMatch() string {
//	files, err := ioutil.ReadDir(".")
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, f := range files {
//		if strings.Contains(f.Name(), ".mp3") {
//			return f.Name()
//		}
//	}
//
//	return ""
//}

func main() {
	r := gin.Default()
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

				c.JSON(200, strings.Replace(file, "\"", "", -1))

				e := os.Remove(file)
				if e != nil {
					c.JSON(500, gin.H{
						"message": e.Error(),
					})

					return
				}
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
	r.GET("/check", func(c *gin.Context) {
		//urlPath := c.Query("q")
		//if len(urlPath) != 11 {
		//	c.JSON(500, "invalid URL")
		//	return
		//}
		//url := "https://www.googleapis.com/youtube/v3/videos?id=" + urlPath + "&part=contentDetails&key=AIzaSyAfAI5KU0Cmh6oKOeAjXskv4yMfc4Xzg8k"



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