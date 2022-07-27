package main

import (
	"fmt"
	"github.com/aldelo/common/wrapper/aws/awsregion"
	"github.com/aldelo/common/wrapper/s3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// "github.com/aldelo/common/wrapper/aws/awsregion"
//	"github.com/aldelo/common/wrapper/s3"

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func handler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	urlPath := r.FormValue("q")

	if len(urlPath) == 11 && urlPath != "favicon.ico" {
		log.Printf("Got request for : %s", urlPath)
		//err := exec.Command("youtube-dl", "--extract-audio", "--audio-format", "mp3", "--output", r.URL.Path[1:] + ".%%(ext)s", r.URL.Path[1:]).Run()
		err := exec.Command("youtube-dl", "--extract-audio", "--audio-format", "mp3", "--output", "%(title)s.%(ext)s", "--restrict-filenames", urlPath).Run()
		if err != nil {
			log.Printf("Error occurred processing URL : %s", urlPath)
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
				log.Fatal("could not connect to s3 - aws")
			}
			f, err := os.Open(file)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			byteContainer, err := ioutil.ReadAll(f)
			if err != nil {
				log.Fatal(err)
			}

			location, err := s.Upload(nil, byteContainer, "music/" + urlPath+".mp3")
			if err != nil {
				log.Fatal(err)
			}

			log.Println(location)

			w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
			w.Header().Set("Content-Type", "audio/mp3")
			//http.ServeFile(w, r, r.URL.Path[1:]+".mp3")
			// return file (or the title of the video)
			_, _ = fmt.Fprintf(w, file)

			e := os.Remove(file)
			if e != nil {
				log.Fatal(e)
			}
		}

	} else{
		log.Printf("Bad URL : %s", urlPath)
	}
}
func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/upload", handler)
	mux.HandleFunc("/health", healthHandler)

	log.Printf("Youtube MP3 Download Backend Server Started")
	//handler := cors.Default().Handler(mux)
	_ = http.ListenAndServe(":8080", mux)
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
//  GOOS=linux GOARCH=amd64 go build -o lysten_api_linux
// ssh -i ~/.ssh/spacedev.pem ubuntu@54.197.68.232
// scp -i ~/.ssh/spacedev.pem lysten_api_linux ubuntu@54.197.68.232:/home/ubuntu/lystenapi
// systemctl --lines=5000 status lystenapi