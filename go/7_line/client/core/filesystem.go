package core

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"line/client/log"
	"line/client/model"
	"line/common"
	"line/grpcchannel"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type fsServer struct {
	server *http.Server
	port2  string
}

var filesystemServer fsServer
var winDiskIndex = 0
var winSlash string

func handleFilesystem(pong *grpcchannel.Pong, cc grpcchannel.ChannelClient) {
	var arg grpcchannel.RPxyResp
	err := json.Unmarshal(pong.Data, &arg)
	if err != nil {
		log.Errorf("handleFilesystem: Unmarshal json:%v\n", err)
		return
	}

	port := strconv.Itoa(int(common.RandomAvaliblePort()))
	addr3 := "127.0.0.1:" + port
	webRoot := "/"
	if runtime.GOOS == "windows" {
		webRoot = model.WinDiskList[winDiskIndex] + "/"
		winDiskIndex = (winDiskIndex + 1) % len(model.WinDiskList)
	}
	go serveFilesystem(addr3, webRoot)

	handleRPxy(pong, cc, addr3)
}

var errTemplate *template.Template
var dirTemplate *template.Template

func init() {
	if runtime.GOOS == "windows" {
		winSlash = "/"
	}

	var err error
	errTemplate, err = template.New("errTemp").Parse(HTML_ERR)
	errFatal(err)

	dirNameFunc := func(path string) string {
		path = filepath.ToSlash(filepath.Dir(strings.TrimRight(path, "/")))
		if strings.HasSuffix(path, "/") {
			return path
		}
		return path + "/"
	}

	dirTemplate, err = template.New("dirTemp").Funcs(template.FuncMap{"dirName": dirNameFunc}).Parse(HTML_DIR)
	errFatal(err)
}

func serveFilesystem(addr, rootDir string) {
	if runtime.GOOS == "windows" {
		ok, err := regexp.MatchString(`^[a-zA-Z]:$`, rootDir)
		if err == nil && ok {
			rootDir = rootDir + `/`
		}
	}
	rootDir = filepath.ToSlash(rootDir)

	m := http.NewServeMux()
	m.Handle("/", http.HandlerFunc(fs(rootDir)))
	m.Handle("/favicon.ico", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) { wr.Write(favicon) }))
	filesystemServer.server = &http.Server{
		Handler: m,
		//go 1.11的http.Server没有BaseContext字段
		//BaseContext: func(ln net.Listener) context.Context {
		//	fileSystemListener.listener = ln
		//	return context.Background()
		//},
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	err = filesystemServer.server.Serve(ln)
	if err != nil {
		log.Errorf("Server.Serves()失败,%v\n", err)
		return
	}
}

func fs(rootDir string) func(wr http.ResponseWriter, req *http.Request) {
	return func(wr http.ResponseWriter, req *http.Request) {
		wr.Header().Set("Cache-Control", "no-store")

		path, err := getPath(req.URL.Path, rootDir)
		if err != nil {
			renderHTMLErr(wr, err.Error())
			return
		}

		switch req.Method {
		case "GET":
			listFS(wr, req, path)
			return
		case "POST":
			uploadFiles(wr, req, path)
			return
		default:
			http.Error(wr, "不支持的方法", http.StatusMethodNotAllowed)
			return
		}
	}
}

var homeURL string

func listFS(wr http.ResponseWriter, req *http.Request, path string) {
	if req.URL.Path == "/" {
		home, _ := url.QueryUnescape(req.FormValue("home"))
		if home != "" {
			homeURL = home
		}
	}

	fi, err := os.Stat(path)
	if err != nil {
		renderHTMLErr(wr, err.Error())
		return
	}

	if fi.Mode().IsDir() {
		hf, err := http.Dir(path).Open("")
		if err != nil {
			renderHTMLErr(wr, err.Error())
			return
		}
		defer hf.Close()

		fis, _ := hf.Readdir(-1) //这里忽略err是为了把能列出的文件/目录列出来
		renderHTMLDir(wr, path, fis)
		return
	}

	if fi.Mode().IsRegular() {
		log.Infof("%q 下载了 %q, %d字节\n", clientIP(req.RemoteAddr), path, fi.Size())
		http.ServeFile(wr, req, path)
		return
	}

	renderHTMLErr(wr, "路径不存在或访问的路径不是目录/常规文件")
}

func uploadFiles(wr http.ResponseWriter, req *http.Request, path string) {
	begin := time.Now()
	err := os.Chdir(path)
	errFatal(err)

	req.ParseMultipartForm(64 << 20) //64MB内存buffer
	var uplFail, upSucc int
	var rename string

	if req.MultipartForm == nil {
		renderHTMLErr(wr, "req.MultipartForm == nil")
		return
	}

	var totalSize int64
	for _, fileHeader := range req.MultipartForm.File["uploadFiles"] {
		srcFile, err := fileHeader.Open()
		if err != nil {
			renderHTMLErr(wr, fmt.Sprintf("打开文件(%s)失败: %v", fileHeader.Filename, err))
			return
		}

		fname := fileHeader.Filename
		upflag := "-上传"
		//循环检查上传文件是否和服务端文件重名，如果文件存在，则重命名上传文件,也就是“.扩展名” 前加 "upflagN"，加了后还重名就继续加
		for i := 1; ; i++ {
			_, err := os.Stat(fname)
			if err != nil {
				break
			}

			suffix := filepath.Ext(fname)
			withoutSuf := strings.TrimSuffix(fname, suffix)
			j := strings.LastIndex(withoutSuf, upflag)
			if j <= 0 {
				fname = withoutSuf + fmt.Sprintf("%s%d", upflag, i) + suffix
				continue
			}

			ind, err := strconv.Atoi(withoutSuf[j+len(upflag):])
			if err != nil {
				fname = withoutSuf + fmt.Sprintf("%s%d", upflag, i) + suffix
				continue
			}
			ind++
			i = ind
			fname = withoutSuf[:j] + fmt.Sprintf("%s%d", upflag, ind) + suffix
		}

		if fileHeader.Filename != fname {
			rename = rename + fmt.Sprintf("%-s&nbsp&nbsp<b>上传文件重名，上传文件被重命名为:</b>&nbsp&nbsp%-s <br />", fileHeader.Filename, fname)
		}

		dstFile, err := os.Create(fname) //创建上传文件
		if err != nil {
			renderHTMLErr(wr, "uploadFiles:os.Create:"+err.Error())
			srcFile.Close()
			return
		}

		n, err := io.Copy(dstFile, srcFile)
		if err != nil {
			uplFail++
		} else {
			log.Infof("%q 上传 %q 成功, %d字节\n", clientIP(req.RemoteAddr), fname, n)
			upSucc++
			totalSize += n
		}
		srcFile.Close() //这里是循环，避免用defer *.Close()
		dstFile.Close()
	}

	var dur = float64(time.Now().Sub(begin)) / float64(time.Second)
	var totalMB = float64(totalSize) / 1024.0 / 1024.0
	var speed = totalMB / dur
	log.Infof("平均速率:%.2f MB/s, 耗时%.2fs, 总大小%.2f MB  %d\n", speed, dur, totalMB, time.Now().Sub(begin))

	fmt.Fprintf(wr, UPLOAD_STATISTIC, "/"+path, 3+len(strings.Split(rename, "<br />")), speed, dur, totalMB, uplFail, upSucc, rename)
}

func clientIP(remoteAddr string) string {
	return remoteAddr[:strings.LastIndex(remoteAddr, ":")]
}

func renderHTMLErr(wr io.Writer, errStr string) {
	err := errTemplate.Execute(wr, errStr)
	errFatal(err)
}

type fsList struct {
	Home  string
	Path  string
	Dirs  []os.FileInfo
	Files []os.FileInfo
}

type fiList []os.FileInfo

func (fl fiList) Len() int {
	return len(fl)
}

func (fl fiList) Less(i, j int) bool {
	return strings.ToLower(fl[i].Name()) < strings.ToLower(fl[j].Name())
}

func (fl fiList) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func renderHTMLDir(wr io.Writer, path string, fis []os.FileInfo) {
	fs := make([]os.FileInfo, len(fis))
	ds := make([]os.FileInfo, len(fis))
	f := 0
	d := 0

	//只显示常规文件和目录
	for i := range fis {
		if fis[i].Mode().IsRegular() {
			fs[f] = fis[i]
			f++
		}
		if fis[i].Mode().IsDir() {
			ds[d] = fis[i]
			d++
		}
	}

	var fl fsList
	fl.Home = homeURL
	fl.Path = winSlash + filepath.ToSlash(path)
	if fl.Path == "/" {
		fl.Path = "." //修复根目录作为web root时url不可用的bug
	}
	fl.Files = fs[:f]
	fl.Dirs = ds[:d]

	//排序,目录在前,文件在后
	sort.Sort(fiList(fl.Files))
	sort.Sort(fiList(fl.Dirs))
	err := dirTemplate.Execute(wr, fl)
	errFatal(err)
}

func getPath(urlPath, rootDir string) (string, error) {
	if urlPath == "/" {
		urlPath = rootDir
	}

	if runtime.GOOS == "windows" && strings.HasPrefix(urlPath, "/"+rootDir) {
		urlPath = urlPath[1:]
	}

	urlPath = filepath.ToSlash(filepath.Clean(urlPath))

	if !strings.HasPrefix(urlPath, rootDir) {
		return "", fmt.Errorf("无权访问%s", urlPath)
	}
	return urlPath, nil
}

func errFatal(err error) {
	if err != nil {
		panic(err)
	}
}

const (
	HTML_ERR = `<!doctype html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <meta http-equiv="Refresh" content="3; url=/">
    <title>Web文件管理</title>
</head>
<body>
<strong>{{.}}</strong>
<br />
<p>正在跳转到根目录...</p>
</body>
</html>
`

	UPLOAD_STATISTIC = `
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
</head>
<body onload="setTimeout(&quot;location.href='%s'&quot;, %d000);">
    <h1>平均速率: %.2fMB/s,  耗时:%.2fs,  总大小: %.2fMB,  上传失败:%d, 成功:%d, </h1> <p>%s</p>
</body>
</html>
`

	HTML_DIR = `<!doctype html>
<html lang="zh">
<head>
    {{ $data := . -}}
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>WEB文件管理</title>
	<style type="text/css">
        form{
            background-color: #EEEEEE;
            position: relative;
            border: 1px solid gray;
            border-radius: 0.2em;
            width:332px;
        }
        #上传按钮{
            position: absolute;
            float: right;
        }
        #文件表格{
            width: 100%;
            border-collapse: collapse;
        }
        #文件表格 tr:nth-child(even){
            background-color: #EEE;
        }
        #文件表格 td:nth-child(odd){text-align: left;}
        #文件表格 td:nth-child(even){text-align: right;}
        #文件表格 td>a{text-decoration:none; }
    </style>
</head>
<body>
{{ $data := . -}}
<header>
    <form enctype="multipart/form-data" action="{{$data.Path}}" method="POST">
      <abbr title="可以按Ctrl键选择多个文件">
          <input type="file" multiple name="uploadFiles" required>
          <input id="上传按钮" type="submit" value="批量上传文件">
      </abbr>
  </form>
  <br />
  <a href="{{$data.Home}}"><b>&#8634; 返回主机管理界面</b></a><br />
  <a href="/"><b>&#8634; 返回web根目录</b></a><br />
  <a href="{{dirName $data.Path}}"><b>&#8634; 返回上层目录</b></a>
  <div style="color: #104E8B"><span style="font-weight: bold">当前目录:</span> {{$data.Path}}</div>
</header>

<article>
    <hr>
    <table id="文件表格">
        <thead style="background-color: #EEEEFF;"><th style="text-align:left;">目录名/文件名</th><th style="text-align:right">大小</th></thead>
        <tbody>
        {{- range $index,$dir := $data.Dirs -}}
            <tr>
                <td class="col1"><a href="{{$data.Path}}/{{$dir.Name}}/"  title="点击打开目录">&bull; {{$dir.Name}}/</a></td>
                <td class="col2">{{$dir.Size}}</td>
            </tr>
        {{end}}
        {{- range $index,$file := $data.Files -}}
            <tr>
                <td class="col1"><a href="{{$data.Path}}/{{$file.Name}}" title="下载纯文本文件: 右键->链接另存为">&bull; {{$file.Name}}</a></td>
                <td class="col2">{{$file.Size}}</td>
            </tr>
        {{end}}
        </tbody>
    </table>
    <hr>
</article>
</body>
</html>
`
)

var favicon = []byte{82, 73, 70, 70, 12, 1, 0, 0, 87, 69, 66, 80, 86, 80, 56, 88, 10, 0, 0, 0, 16, 0, 0, 0, 15, 0, 0, 15, 0, 0, 65, 76, 80, 72, 87, 0, 0, 0, 1, 199, 160, 160, 141, 36, 53, 118, 248, 204, 44, 232, 35, 34, 32, 189, 230, 57, 226, 177, 108, 240, 47, 204, 200, 140, 15, 137, 161, 182, 141, 36, 229, 238, 153, 153, 250, 175, 149, 162, 207, 35, 250, 63, 1, 192, 124, 206, 105, 155, 179, 103, 64, 226, 186, 142, 68, 245, 187, 174, 236, 188, 155, 214, 189, 184, 255, 152, 113, 55, 12, 47, 119, 133, 163, 14, 181, 29, 107, 208, 8, 72, 133, 60, 10, 138, 0, 0, 86, 80, 56, 32, 142, 0, 0, 0, 80, 2, 0, 157, 1, 42, 16, 0, 16, 0, 2, 0, 52, 37, 176, 2, 116, 6, 46, 191, 7, 153, 15, 201, 43, 192, 64, 0, 254, 215, 63, 112, 129, 255, 234, 119, 179, 93, 185, 78, 153, 182, 9, 86, 217, 37, 171, 172, 11, 215, 181, 23, 149, 27, 223, 20, 190, 93, 57, 166, 107, 82, 13, 31, 151, 53, 137, 113, 238, 121, 1, 162, 219, 215, 79, 231, 172, 194, 48, 242, 108, 203, 237, 193, 32, 198, 214, 240, 209, 26, 245, 135, 224, 57, 249, 158, 162, 253, 99, 230, 170, 155, 7, 249, 143, 127, 59, 56, 119, 198, 21, 252, 212, 90, 90, 51, 53, 215, 255, 18, 42, 186, 199, 175, 103, 139, 250, 40, 24, 169, 117, 128, 121, 117, 185, 111, 65, 101, 242, 64, 232, 0}
