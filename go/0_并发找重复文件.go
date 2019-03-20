/* Description： 找出参数1指定的目录下的重复文件，如果没有制定参数1，则找出当前目录下的重复文件
*  Author: Zhouhui
*  Release-date: 2017-12-26
 */
package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
)

var openMax uint64

func main() {
	var rootDir string
	var e error
	if len(os.Args) < 2 {
		rootDir, e = os.Getwd()
		if e != nil {
			log.Fatal(e)
		}
	} else {
		rootDir = os.Args[1]
	}

	var rlim syscall.Rlimit
	if e = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim); e != nil {
		log.Fatalf("get max open fd of process failed, %v", e)
	}
	openMax = rlim.Cur //获取进程最大打开文件句柄数目限制

	md5Name, e := findDup(rootDir)
	if e != nil {
		log.Fatal(e)
	}

	var sameMD5Files []string
	for _, files := range md5Name {
		sort.Strings(files)
		sameMD5Files = append(sameMD5Files, strings.Join(files, "\n"))
	}

	// 按路径排序并打印重复文件路径到stdout
	sort.Strings(sameMD5Files)
	for _, f := range sameMD5Files {
		fmt.Printf("@@@@@\n%s\n", f)
	}
}

//IN: 目录路径
//OUT1: 映射(md5->文件列表)
//OUT2: error
func findDup(dir string) (map[string][]string, error) {
	szName := make(map[int64][]string)   //映射：文件大小 -> 文件列表
	md5Name := make(map[string][]string) //映射：文件MD5 -> 文件列表

	if e := walkdir(dir, szName); e != nil {
		return nil, e
	}

	sumCh := make(chan map[string]string, openMax*8/10)
	var wg sync.WaitGroup
	tokens := make(chan struct{}, openMax*8/10) //限制最大goroutine并发数
	for size, name := range szName {
		if len(name) < 2 {
			delete(szName, size)
			continue
		}

		for _, n := range name {
			wg.Add(1)
			go func(file string) {
				defer wg.Done()
				defer func() { <-tokens }()
				tokens <- struct{}{}

				f, e := os.Open(file)
				if e != nil {
					log.Printf("Open, %v\n", e)
					return
				}
				defer f.Close()

				m5 := md5.New()
				io.Copy(m5, f)
				sumCh <- map[string]string{fmt.Sprintf("%x", m5.Sum(nil)): file}
			}(n)
		}

	}

	go func() {
		wg.Wait()
		close(sumCh)
	}()

	for sum := range sumCh {
		for k, v := range sum {
			md5Name[k] = append(md5Name[k], v)
		}
	}

	for md5, files := range md5Name {
		if len(files) < 2 {
			delete(md5Name, md5)
			continue
		}

	}

	return md5Name, nil
}

//IN: dir 目录路径
//OUT1: szName 入参形式的输出，映射(文件大小->文件列表)
//OUT2: error
func walkdir(dir string, szName map[int64][]string) error {
	var fis []os.FileInfo
	fis, e := ioutil.ReadDir(dir)
	if e != nil {
		return e
	}

	for _, fi := range fis {
		switch {
		case fi.IsDir():
			if e := walkdir(filepath.Join(dir, fi.Name()), szName); e != nil {
				return e
			}
		case fi.Mode().IsRegular():
			if fi.Size() == 0 {
				continue
			}
			sz := fi.Size()
			szName[sz] = append(szName[sz], filepath.Join(dir, fi.Name()))
		}
	}

	return nil
}
