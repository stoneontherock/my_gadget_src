//Release: 2019-07-09
package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var addr *string
var rootDir *string

var errTemplate *template.Template
var dirTemplate *template.Template

func init() {
	binDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	addr = flag.String("a", ":80", "http监听地址,例如: ':8000'或'192.168.0.100:8000'")
	rootDir = flag.String("d", binDir, "http监听地址,例如: ':8000'或'192.168.0.100:8000'")
	flag.Parse()

	var err error
	*rootDir, err = filepath.Abs(*rootDir)
	errFatal(err)

	errTemplate, err = template.New("errTemp").Parse(HTML_ERR)
	errFatal(err)

	dirNameFunc := func(path string) string {
		path = strings.TrimRight(path, "/")
		return filepath.Dir(path) + "/"
	}
	dirTemplate, err = template.New("dirTemp").Funcs(template.FuncMap{"dirName": dirNameFunc}).Parse(HTML_DIR)
	errFatal(err)
}

func main() {
	log.Printf("当前监听地址:%s  当前web根目录:%s", *addr, *rootDir)
	log.Println("通过-a命令行选项修改监听地址, 例如: -a 192.168.1.100:8000 或 -a :8000")
	log.Println("通过-d命令行选项修改WEB根目录, 例如: -d /tmp")
	println()
	serve()
}

func serve() {
	http.HandleFunc("/", fs)
	http.HandleFunc("/favicon.ico", func(wr http.ResponseWriter, req *http.Request) { wr.Write(favicon) })
	err := http.ListenAndServe(*addr, nil)
	errFatal(err)
}

func fs(wr http.ResponseWriter, req *http.Request) {
	path, err := getPath(req.URL.Path)
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
	default:
		http.Error(wr, "不支持的方法", http.StatusMethodNotAllowed)
		return
	}

}

func listFS(wr http.ResponseWriter, req *http.Request, path string) {
	fi, err := os.Stat(path)
	if err != nil {
		renderHTMLErr(wr, err.Error())
		return
	}

	if fi.Mode().IsDir() {
		hf, err := http.Dir(path).Open("")
		errFatal(err)
		fis, err := hf.Readdir(-1)
		errFatal(err)

		renderHTMLDir(wr, path, fis)
		return
	}

	if fi.Mode().IsRegular() {
		http.ServeFile(wr, req, path)
		return
	}

	renderHTMLErr(wr, "路径不存在或访问的路径不是目录/常规文件")
}

func uploadFiles(wr http.ResponseWriter, req *http.Request, path string) {
	err := os.Chdir(path)
	errFatal(err)

	req.ParseMultipartForm(64 << 20) //64MB内存buffer
	var uplFail, upSucc int
	var rename string

	if req.MultipartForm == nil {
		renderHTMLErr(wr, "req.MultipartForm == nil")
		return
	}

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
			rename = rename + fmt.Sprintf("%-s&nbsp&nbsp<b>上传文件重名，上传文件被重命名为:</b>&nbsp&nbsp%-s </br>", fileHeader.Filename, fname)
		}

		dstFile, err := os.Create(fname) //创建上传文件
		if err != nil {
			renderHTMLErr(wr, "uploadFiles:os.Create:"+err.Error())
			srcFile.Close()
			return
		}

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			uplFail++
		} else {
			log.Printf("%s upload %q success\n", req.RemoteAddr[:strings.LastIndex(req.RemoteAddr, ":")], fname)
			upSucc++
		}
		srcFile.Close() //这里是循环，避免用defer *.Close()
		dstFile.Close()
	}

	fmt.Fprintf(wr, "<h1>上传失败:%d, 成功:%d</h1> <p>%s</p> <script language='javascript' type='text/javascript'> setTimeout(\"javascript:location.href='%s'\", %d000); </script>",
		uplFail, upSucc, rename, path, 1+len(strings.Split(rename, "</br>")))
}

func renderHTMLErr(wr io.Writer, errStr string) {
	err := errTemplate.Execute(wr, errStr)
	errFatal(err)
}

type fsList struct {
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
	fl.Path = path
	fl.Files = fs[:f]
	fl.Dirs = ds[:d]

	//排序,目录在前,文件在后
	sort.Sort(fiList(fl.Files))
	sort.Sort(fiList(fl.Dirs))
	err := dirTemplate.Execute(wr, fl)
	errFatal(err)
}

func getPath(path string) (string, error) {
	p := path
	if p == "/" {
		p = *rootDir
	}

	if !strings.HasPrefix(p, *rootDir) {
		return "", fmt.Errorf("不能访问%s", p)
	}
	return p, nil
}

func errFatal(err error) {
	if err != nil {
		panic(err)
	}
}

const (
	HTML_ERR = `
<!doctype html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Web文件管理</title>
</head>
<body>
<strong>{{.}}<br />正在跳转到根目录...</strong>
<script language='javascript' type='text/javascript'> setTimeout("javascript:location.href='/'", 3000); </script>
</body>
</html>
`

	HTML_DIR = `
<!doctype html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <style type="text/css">
        #上传form{
            height: 2em;
        }
        #文件表格{
            width: 100%;
            border-collapse: collapse;
        }
        tr:nth-child(even){
            background-color: #EEE;
        }
        td.col1{text-align: left}
        td.col2{text-align: right}
        a.文件列表{text-decoration:none; }
    </style>
    <title>WEB文件管理</title>
</head>
<body>
{{- $data := . -}}
<header>
    <form id="上传form" enctype="multipart/form-data" action="{{$data.Path}}" method="POST">
        <input type="file" multiple name="uploadFiles"/>
        <input type="submit" value="批量上传" />
    </form>
    <a class="a返回" href="/"  class="name"><b>&#8634; 返回根目录</b></a><br />
    <a class="a返回" href="{{dirName $data.Path}}"  class="name"><b>&#8634; 返回上层目录</b></a>
</header>

<article>
    <br />
    <table id="文件表格">
        <thead><th style="border-bottom:1px dotted red; text-align:left">文件/目录</th><th style="border-bottom:1px dotted red; text-align:right">大小</th></thead>
        <tbody>
        {{- range $index,$dir := $data.Dirs -}}
            <tr>
                <td class="col1"><a href="{{$data.Path}}/{{$dir.Name}}/"  title="点击打开目录" class="文件列表">&bull; {{$dir.Name}}/</a></td>
                <td class="col2">{{$dir.Size}}</td>
            </tr>
        {{end}}
        {{- range $index,$file := $data.Files -}}
            <tr>
                <td class="col1"><a href="{{$data.Path}}/{{$file.Name}}"  class="文件列表">&bull; {{$file.Name}}</a></td>
                <td class="col2">{{$file.Size}}</td>
            </tr>
        {{end}}
        </tbody>
    </table>
</article>
</body>
</html>
`
)

var favicon = []byte{82, 73, 70, 70, 12, 1, 0, 0, 87, 69, 66, 80, 86, 80, 56, 88, 10, 0, 0, 0, 16, 0, 0, 0, 15, 0, 0, 15, 0, 0, 65, 76, 80, 72, 87, 0, 0, 0, 1, 199, 160, 160, 141, 36, 53, 118, 248, 204, 44, 232, 35, 34, 32, 189, 230, 57, 226, 177, 108, 240, 47, 204, 200, 140, 15, 137, 161, 182, 141, 36, 229, 238, 153, 153, 250, 175, 149, 162, 207, 35, 250, 63, 1, 192, 124, 206, 105, 155, 179, 103, 64, 226, 186, 142, 68, 245, 187, 174, 236, 188, 155, 214, 189, 184, 255, 152, 113, 55, 12, 47, 119, 133, 163, 14, 181, 29, 107, 208, 8, 72, 133, 60, 10, 138, 0, 0, 86, 80, 56, 32, 142, 0, 0, 0, 80, 2, 0, 157, 1, 42, 16, 0, 16, 0, 2, 0, 52, 37, 176, 2, 116, 6, 46, 191, 7, 153, 15, 201, 43, 192, 64, 0, 254, 215, 63, 112, 129, 255, 234, 119, 179, 93, 185, 78, 153, 182, 9, 86, 217, 37, 171, 172, 11, 215, 181, 23, 149, 27, 223, 20, 190, 93, 57, 166, 107, 82, 13, 31, 151, 53, 137, 113, 238, 121, 1, 162, 219, 215, 79, 231, 172, 194, 48, 242, 108, 203, 237, 193, 32, 198, 214, 240, 209, 26, 245, 135, 224, 57, 249, 158, 162, 253, 99, 230, 170, 155, 7, 249, 143, 127, 59, 56, 119, 198, 21, 252, 212, 90, 90, 51, 53, 215, 255, 18, 42, 186, 199, 175, 103, 139, 250, 40, 24, 169, 117, 128, 121, 117, 185, 111, 65, 101, 242, 64, 232, 0}
