//功能： 读取重复文件记录文件，删除重复文件
//参数： 1.重复文件记录文件, 2.分割字符串
/*
  重复文件记录文件格式类似(假设分隔字符串为"@@@@@"):
      @@@@@
      /tmp/a.c
      /tmp/b.c
      @@@@@
      /tmp/11.jpg
      /tmp/11_(1).jpg
*/
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: %s <file> <split string>")
		os.Exit(1)
	}

	f, e := os.Open(os.Args[1])
	if e != nil {
		log.Fatal(e)
	}

	bytes, e := ioutil.ReadAll(f)
	if e != nil && e != io.EOF {
		log.Fatal(e)
	}

	for _, files := range strings.Split(string(bytes), os.Args[2]) {
		var fileS []string
		for _, s := range strings.Split(files, "\n") {
			if s != "" {
				fileS = append(fileS, s)
			}
		}

		if len(fileS) < 2 {
			continue
		}

		sort.Strings(fileS)
		prefix := dirPrefix(fileS)
		fmt.Printf("\n目录:\"%s/\"\n", prefix)
		for i, f := range fileS {
			fmt.Printf("%4d: %s\n", i+1, strings.TrimPrefix(f, prefix+"/"))
		}
		fmt.Printf("你要留哪个？ (全部留下则直接回车,全删按0):")

		var index int
		cnt, _ := fmt.Fscanf(os.Stdin, "%d", &index)
		if cnt == 0 {
			continue
		}

		for i, f := range fileS {
			if i == index-1 {
				continue
			}
			fmt.Printf(">>> 删除 %s\n", f)
			os.Remove(f)
		}

	}
}

func dirPrefix(path []string) string {
	newPath := make([]string, len(path))
	copy(newPath, path)

	for ind := len(newPath) - 1; ind > 0; ind-- {
		p1 := strings.Split(filepath.Clean(newPath[ind-1])+"/", "/")
		p2 := strings.Split(filepath.Clean(newPath[ind])+"/", "/")
		if len(p1) > len(p2) {
			p1, p2 = p2, p1
		}

		var offset int
		for i, s := range p1 {
			if p2[i] != s {
				offset = i
				break
			}
		}
		newPath[ind-1] = strings.Join(p1[:offset], "/")
	}
	return newPath[0]
}
