package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kr/pretty"
	"github.com/remeh/sizedwaitgroup"
)

func do_file(filename string) []string {
	file_bytes, _ := ioutil.ReadFile(filename)
	data := strings.Split(string(file_bytes), "\n")
	return data
}

func main() {
	start := time.Now()
	file_path := flag.String("dir", "/data/file.txt", "path to directory with line protocol files")
	token := flag.String("token", "W6V6XI22jGMuZmTymK2vkG---yYa912v0xa_83h7VxMiZLijdKlNfyQxho01TI4KP1L4oEAWir7eu3tFFetL-Q==", "influx auth token")
	bucket := flag.String("bucket", "gpu_monitoring", "bucket name")
	organisation := flag.String("org", "Lancaster University", "organisation name")
	server := flag.String("server", "http://localhost:9999", "host and port of server")

	flag.Parse()

	var files []string
	filepath.Walk(*file_path, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	// pre-process file data and load it
	var files_data []string
	var queue chan []string = make(chan []string, 1)
	var wg sync.WaitGroup
	wg.Add(len(files))
	for i := 0; i < len(files); i++ {
		go func(i int) {
			queue <- do_file(files[i])
		}(i)
	}
	go func() {
		for d := range queue {
			files_data = append(files_data, d...)
			wg.Done()
		}
	}()
	wg.Wait()

	fmt.Println(len(files_data))

	client := &http.Client{}

	swg := sizedwaitgroup.New(30)

	var l int = len(files_data)

	for i := 0; i < l; i++ {
		swg.Add()
		go func(i int) {
			defer swg.Done()
			send(i, l, *server, *token, *organisation, *bucket, client, files_data[i])
		}(i)
	}

	swg.Wait()
	elapsed := time.Since(start)
	log.Printf("Took %s", elapsed)
}

func send(idx int, l int, server string, token string, organisation string, bucket string, client *http.Client, text string) {
	if idx != 0 && idx%1000 == 0 {
		fmt.Printf("%d/%d\n", idx, l)
	}
	if text == "" {
		return
	}
	r, err := compressWithGzip(strings.NewReader(text))
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("POST", server+"/api/v2/write", r)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Authorization", "Token "+token)
	q := req.URL.Query()
	q.Add("org", organisation)
	q.Add("bucket", bucket)
	q.Add("precision", "ns")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 204 {
		bytes, _ := ioutil.ReadAll(resp.Body)
		pretty.Println(string(bytes))
		pretty.Println(text)
	}
}

func compressWithGzip(data io.Reader) (io.Reader, error) {
	pr, pw := io.Pipe()
	gw := gzip.NewWriter(pw)
	var err error
	go func() {
		_, err = io.Copy(gw, data)
		gw.Close()
		pw.Close()
	}()
	return pr, err
}
