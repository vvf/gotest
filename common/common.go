package common

import (
	"encoding/binary"
	"io"
	"bufio"
)


type FileStat struct {
	Stat     [256]uint32
	Filename string

}

func (self *FileStat) Inc(i byte){
	self.Stat[i] ++
}

func (self *FileStat) Bytes() []byte{
	// TODO: use gob package
	packet := make([]byte, 1026 + len(self.Filename) ) // last two bytes should by zero
	for i, v := range self.Stat {
		binary.LittleEndian.PutUint32(packet[i*4:(i+1)*4], v)
	}
	copy(packet[1024:], self.Filename)
	return packet
}

func (self *FileStat) Read(c io.Reader) error {
	// TODO: use gob package
	buf_reader := bufio.NewReader(c)

	for i:=0; i<256;i++ {
		dat :=make([]byte, 4)
		_, err := io.ReadFull(c, dat)
		if err != nil {
			return err
		}
		self.Stat[i] = binary.LittleEndian.Uint32(dat)
	}
	fname, err := buf_reader.ReadString(0)
	self.Filename = fname
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
