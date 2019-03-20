//安全删除目录或文件(防止被数据恢复软件恢复)

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Usage: %s <path>")
	}

	fi, err := os.Stat(os.Args[1])
	if err != nil {
		log.Fatalf("[os.Stat]%v", err)
	}

	var pathSize = make(map[string]int64)
	if fi.IsDir() {
		err = walkdirAndZeroFile(os.Args[1], pathSize)
		if err != nil {
			log.Fatalf("[walkdirAndZeroFile]%v", err)
		}
	} else {
		pathSize[fi.Name()] = fi.Size()
	}

	for path, size := range pathSize {
		err = fillZero(path, size)
		if err != nil {
			log.Fatalf("[fillZero]%v", err)
		}
	}

	err = os.RemoveAll(os.Args[1])
	if err != nil {
		log.Fatalf("%v or some files or dirs in %q are opened.", err, os.Args[1])
	}
}

func fillZero(path string, size int64) error {
	buf := make([]byte, size)
	err := ioutil.WriteFile(path, buf, 0666)
	if err != nil {
		return (err)
	}
	return nil
}

func walkdirAndZeroFile(dir string, pathSize map[string]int64) error {
	infs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fi := range infs {
		if fi.IsDir() {
			walkdirAndZeroFile(filepath.Join(dir, fi.Name()), pathSize)
		} else {
			pathSize[filepath.Join(dir, fi.Name())] = fi.Size()
		}
	}
	return nil
}
