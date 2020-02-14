package core

//
//import (
//	"connekts/client/log"
//	"connekts/grpcchannel"
//	"context"
//	"io/ioutil"
//	"os"
//	"path/filepath"
//	"time"
//)
//
//
//func handleListFile(pong *grpcchannel.Pong, cc grpcchannel.ChannelClient) {
//	pth:= string(pong.Data)
//	if pth == "" {
//		pth, _ = filepath.Abs(".")
//	}
//	println("path:", pth)
//
//	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*60)
//	defer cancel()
//
//	var fl *grpcchannel.FileList
//	fl = lsDir(pth)
//	fl.Mid = staticInfo.MachineID
//	fl.Path = pth
//
//	_, err := cc.ListFile(ctx, fl)
//	if err != nil {
//		log.Errorf("cc.ListFile:%v\n", err)
//		return
//	}
//}
//
//
//
////获取目录/文件列表
//func lsDir(dir string) *grpcchannel.FileList {
//	var fsList grpcchannel.FileList
//
//	var fi []os.FileInfo
//	fi, err := ioutil.ReadDir(dir)
//	if err != nil {
//		log.Errorf("ReadDir(),%v\n", err)
//		fsList.Err = err.Error()
//		return &fsList
//	}
//
//	for _, f := range fi {
//		var gf grpcchannel.File
//		if f.IsDir() {
//			gf.Name = f.Name() + "/"
//			gf.Size = int32(f.Size())
//		} else if f.Mode().IsRegular() {
//			gf.Name = f.Name()
//			gf.Size = int32(f.Size())
//		}else{
//			continue
//		}
//
//		fsList.Fs = append(fsList.Fs,&gf)
//	}
//
//	return &fsList
//}
