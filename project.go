package main

import (
	"fmt"
	"path/filepath"
	"os"
	"time"
	"common"
	"net"
	"sync"
)
const (
	BUFFER_SIZE = 1048576
)

var results chan *common.FileStat
var wg sync.WaitGroup

func read_file_stat(file string) {
//	arr := make([]int16,256,256)
	defer wg.Done()
	fmt.Printf("Start calc '%s'\n", file)
	fs := new(common.FileStat)
	fs.Filename = file
	fh, err := os.Open(file)
	if err !=nil {
		return
	}
	defer fh.Close()
	stat, err := fh.Stat()
    if err != nil {
        return
    }
    // read the file
	// "eat" file by small part (not whole file if it is so big)
	file_size := stat.Size()
	var part_size int64 = BUFFER_SIZE
	for file_size >0 {
		if file_size < part_size {
			file_size = part_size
		}
		bs := make([]byte, part_size)
		_, err = fh.Read(bs)
		if err != nil {
			return
		}
		// It can be go-routine but it will increase used memory size
		for _, b := range bs {
			fs.Inc(b)
		}
		file_size -= part_size
	}

//    str := string(bs)
//    fmt.Println(str)
	fmt.Println("Publish result ",file)
	results <- fs
}

func send_packet(packet []byte){
	defer wg.Done()
	conn, err := net.Dial("tcp", "127.0.0.1:8911")
	if err != nil {
		fmt.Println("Error:")
		fmt.Println(err)
		return
	}
	defer conn.Close()
	fmt.Println("Send data ", len(packet), " bytes")
	conn.Write(packet)
	common_stat := new(common.FileStat)
	common_stat.Read(conn)
	fmt.Println("Common stat is: ", common_stat.Filename)
	fmt.Println(common_stat.Stat)
}

func send_results(){
	fmt.Printf("Wait results\n")
	for {
		fs := <- results

		fmt.Printf("Send result of ")
		fmt.Println(fs.Filename)
		packet := fs.Bytes()
//		fmt.Printf(packet)
		wg.Add(1)
		go send_packet( packet )
	}
}


func walking (path string, info os.FileInfo, err error) error {

	if !info.IsDir(){
		defer wg.Add(1)
		go read_file_stat( path )
	}
	return nil
}

func main() {
	results = make(chan *common.FileStat, 2)
	fmt.Printf("\nHello world!\n")
	go send_results()
	filepath.Walk("files-here", walking)
	time.Sleep(1000)
	wg.Wait()
}
