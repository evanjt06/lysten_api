package main

import (
	"fmt"
	"github.com/aldelo/common/wrapper/aws/awsregion"
	"github.com/aldelo/common/wrapper/s3"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if len(r.URL.Path[1:]) == 11 && r.URL.Path[1:] != "favicon.ico" {
		log.Printf("Got request for : %s", r.URL.Path[1:])
		//err := exec.Command("youtube-dl", "--extract-audio", "--audio-format", "mp3", "--output", r.URL.Path[1:] + ".%%(ext)s", r.URL.Path[1:]).Run()
		err := exec.Command("youtube-dl", "--extract-audio", "--audio-format", "mp3", "--output", "%(title)s.%(ext)s", "--restrict-filenames", r.URL.Path[1:]).Run()
		if err != nil {
			log.Printf("Error occurred processing URL : %s", r.URL.Path[1:])
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

			location, err := s.Upload(nil, byteContainer, r.URL.Path[1:]+".mp3")
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
		log.Printf("Bad URL : %s", r.URL.Path[1:])
	}
}
func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)

	log.Printf("Youtube MP3 Download Backend Server Started")
	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":8080", handler)
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