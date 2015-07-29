package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net"
	"log"
//	"io"
	"common"
	"strings"
)
var results chan *common.FileStat
var commonStat *common.FileStat
var commonStatMutex chan int

func to_json(info *common.FileStat) string{
	var json_result string

	json_result = "{\"filepath\":\"" + info.Filename + "\", "
	json_parts := make([]string, 256)
	commonStatMutex <- 1 // It will block next concurrent call
	for i,n := range info.Stat {
		json_parts[i] = fmt.Sprintf("\"0x%02X\":%d", i, n )
		commonStat.Stat[i] += n
//			json_result += fmt.Sprintf("\"0x%02X\":%d,", i, n )
	}
	<- commonStatMutex
	json_result += strings.Join(json_parts,", ")
	json_result += "}"
	fmt.Println( "Common stat:", commonStat.Stat )
	return json_result
}

func to_bson(info *common.FileStat) bson.D{
	var stat_arr = make([]bson.DocElem, 257)
	commonStatMutex <- 1 // It will block next concurrent call
	stat_arr[0] = bson.DocElem{ Name:"filename", Value: info.Filename }
	for i, v := range info.Stat{
		commonStat.Stat[i] += v
		stat_arr[i+1]=bson.DocElem{Name:fmt.Sprintf("0x%02X",i), Value:v}
	}
	<- commonStatMutex
	log.Println("bsoned ", info.Filename)
	return stat_arr
}


func main() {
	results = make(chan *common.FileStat)
	commonStatMutex = make(chan int,1)
	commonStat = new(common.FileStat)
	commonStat.Filename = "all"
	// Listen on TCP port 2000 on all interfaces.
	mongoSession, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:[]string{"localhost"}, Database:"gotest", Username:"gotest", Password:"GOLang123" })

	if err != nil {
		log.Fatalf("CreateSession: %s\n", err)
	}
	mongoSession.SetMode(mgo.Monotonic, true)
	log.Println("Start listen tcp on 8911")
	l, err := net.Listen("tcp", ":8911")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Echo all incoming data.
			defer c.Close()
			info := new(common.FileStat)
			if err := info.Read(c); err != nil {
				log.Println(err)
				return
			}
//			json_to_save := to_json(info)
			item_to_save := to_bson(info)
			commonStatMutex <- 1 // It will block next concurrent call
			conn.Write(commonStat.Bytes())
			<- commonStatMutex
//			fmt.Println(json_to_save)
//				client.Set('a', dat)
			// Shut down the connection.
			go func() {
				sessionCopy := mongoSession.Copy()
				defer sessionCopy.Close()
				// save received data to mongo sessionCopy
				collection := sessionCopy.DB("gotest").C("filestat")
				collection.Insert(item_to_save)
			}()

		}(conn)
	}
}
