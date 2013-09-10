package main

import (
	"flag"
	"fmt"
	"io"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Cmd struct {
	name      string
	modifiers []string
	conn      *s3.S3
	url       *url.URL
	bucket    s3.Bucket
}

func RunCommand(cmd Cmd) {
	switch cmd.name {
	case "ls":
		Ls(cmd)
	case "get":
		Get(cmd)
	case "rm":
		Rm(cmd)
	}
}

func getRegion(region string) aws.Region {
	switch region {
	case "us-east-1":
		return aws.USEast
	case "us-west-1":
		return aws.USWest
	case "us-west-2":
		return aws.USWest2
	}
	return aws.USEast

}

var region string

func init() {
	flag.StringVar(&region, "r", "us-east-1", "Choose your region")
}

func main() {
	flag.Parse()

	cmds := flag.Args()

	if len(cmds) <= 0 {
		fmt.Println("Command required")
		os.Exit(1)
	}

	cmd := Cmd{}
	cmd.name = cmds[0]
	cmd.modifiers = cmds[1:]

	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}

	cmd.conn = s3.New(auth, getRegion(region))
	s3url, err := url.Parse(cmd.modifiers[0])
	cmd.url = s3url
	cmd.bucket = s3.Bucket{cmd.conn, cmd.url.Host}
	RunCommand(cmd)

}

func KeySender(cmd Cmd, key_chan chan s3.Key) {
	limit := 1000
	marker := ""

	path := strings.Replace(cmd.url.Path, "/", "", 1)
	list, err := cmd.bucket.List("", "/", marker, limit)

	if err != nil {
		panic(err)

	}
	for _, item := range list.Contents {
		marker = item.Key
		if path != "" {
			matched, err := filepath.Match(path, item.Key)
			if err != nil {
				panic(err)
			}
			if !matched && path != "" {
				continue
			}
		}
		key_chan <- item

	}

	for list.IsTruncated == true {
		if marker == list.Marker {
			break
		}
		list, err = cmd.bucket.List("", "/", marker, limit)
		for _, item := range list.Contents {
			marker = item.Key

			matched, err := filepath.Match(path, item.Key)
			if err != nil {
				panic(err)
			}

			if !matched && path != "" {
				continue
			}

			key_chan <- item
		}

	}
}

func KeyPrinter(key_chan chan s3.Key, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-time.After(1 * time.Second):
			return
		case key := <-key_chan:
			fmt.Println(key.Key)
		}
	}
}
func Ls(cmd Cmd) {
	key_chan := make(chan s3.Key)
	wg := new(sync.WaitGroup)

	go KeyPrinter(key_chan, wg)
	wg.Add(1)
	KeySender(cmd, key_chan)

	wg.Wait()
}
func Get(cmd Cmd) {
	key_chan := make(chan s3.Key)

	writers := 5
	wg := new(sync.WaitGroup)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go KeyWriter(i, cmd, key_chan, wg)
	}
	KeySender(cmd, key_chan)

	wg.Wait()
}

func Rm(cmd Cmd) {
	key_chan := make(chan s3.Key)

	writers := 5
	wg := new(sync.WaitGroup)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go KeyDeleter(i, cmd, key_chan, wg)
	}
	KeySender(cmd, key_chan)
	wg.Wait()
}
func KeyDeleter(writer_id int, cmd Cmd, key_chan chan s3.Key, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-time.After(1 * time.Second):
			return
		case key := <-key_chan:
			err := cmd.bucket.Del(key.Key)
			if err != nil {
				panic(err)
			}
		}

	}

}
func KeyWriter(writer_id int, cmd Cmd, key_chan chan s3.Key, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-time.After(1 * time.Second):
			return
		case key := <-key_chan:
			data, err := cmd.bucket.GetReader(key.Key)

			file_path := fmt.Sprintf("./%s", key.Key)
			fo, err := os.Create(file_path)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			buf := make([]byte, 1024)

			for {
				n, err := data.Read(buf)
				if err == io.EOF {
					data.Close()
					fmt.Printf("Downloaded file...%v\n", file_path)
					break
				}
				if err != nil {
					panic(err)
				}
				_, err = fo.Write(buf[:n])

				if err != nil {
					panic(err)
				}

			}
			fo.Close()
		}
	}
}
