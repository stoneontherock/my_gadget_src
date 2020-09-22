//Description: web文件管理器。支持多文件批量上传。文件下载，目录浏览，显示文件大小
//Author: Zhouhui
//Release: 2017-12-26
//Patch1: 2018-11-08

package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type baseSize struct {
	Href string
	Base string
	Size int64
}

type uriBaseSize struct {
	Back string
	BS   []baseSize
}

var rootDir string
var serverStart = strconv.FormatInt(time.Now().UnixNano(), 10)

func main() {
	port := flag.String("p", "80", "listen port")
	root := flag.String("d", ".", "root dir path")
	flag.Parse()

	var err error
	err = os.Chdir(*root)
	if err != nil {
		log.Fatalf("Chdir,%v\n", err)
	}

	rootDir, err = filepath.Abs(*root)
	if err != nil {
		log.Fatalf("Abs,%v\n", err)
	}

	log.Printf("Usage: %s [-p <listen_port>]  [-d <root_directory>]\n", filepath.Base(os.Args[0]))
	log.Printf("Example: %s -p 80  -d /tmp\n\n", filepath.Base(os.Args[0]))
	log.Printf("listening port:%s, root directory:%q\n", *port, rootDir)

	http.HandleFunc("/", fileManager)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { return })
	log.Fatalf("ListenAndServe, %v\n", http.ListenAndServe(":"+*port, nil))
}

//浏览目录，或下载常规文件(不允许软连接，套接字，管道文件下载)
func fileManager(w http.ResponseWriter, r *http.Request) {
	peerAddr := r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")] //peerAddr用来记录浏览器IP地址
	reqPath := rootDir + r.URL.Path

	//此cookie是为了解决: server端重启后，上传文件被上传到错误路径的的问题(会上传到http根)
	ck, err := r.Cookie("server_start")
	if err == http.ErrNoCookie || ck.Value != serverStart {
		ts := http.Cookie{Name: "server_start", Path: "/", Value: serverStart}
		http.SetCookie(w, &ts)
	}

	fi, err := os.Lstat(reqPath)
	if err != nil {
		fmt.Fprintf(w, "<h1>路径 %q 不存在, 跳转到根路径 ...</h1> <script language='javascript' type='text/javascript'> setTimeout(\"javascript:location.href='/'\", 2000); </script>", r.URL.Path)
		return
	}

	switch {
	case fi.IsDir():
		//此cookie是为了上传文件在服务器重启后,上传动作终止，跳转到原路径
		lastDir := http.Cookie{Name: "last_visit_dir", Path: "/", Value: r.URL.Path}
		http.SetCookie(w, &lastDir)

		os.Chdir(reqPath)
		genHtml(reqPath, w)
		log.Printf("%s view dir %q\n", peerAddr, reqPath)
		return
	case fi.Mode().IsRegular():
		if fi.Size() == 0 {
			fmt.Fprintf(w, "<h1>文件%q大小为0字节</h1>\n", r.URL.Path)
			return
		}
		fn, _ := url.PathUnescape(reqPath)
		f, err := os.Open(fn)
		if err != nil {
			log.Printf("os.Open, %v\n", err)
			return
		}
		defer f.Close()
		n, err := io.Copy(w, f)
		if err != nil {
			log.Printf("io.Copy, %v\n", err)
			return
		}
		log.Printf("%s download %q (%d Bytes)\n", peerAddr, reqPath, n)
	default:
		fmt.Fprintf(w, "<h1>%q不是目录也不是常规文件</h1>", reqPath)
	}
}

//多文件批量上传
func upload(w http.ResponseWriter, r *http.Request) {
	ck, err := r.Cookie("server_start")
	if err == http.ErrNoCookie || ck.Value != serverStart {
		reqPath := "/"
		lastVisitDir, err := r.Cookie("last_visit_dir")
		if err == nil {
			reqPath = lastVisitDir.Value
		}
		w.WriteHeader(404)
		fmt.Fprintf(w, "<html><h1>因会话过期，上传文件失败。</h1><h1>请在页面跳转后重试...</h1> <script language='javascript' type='text/javascript'> setTimeout(\"javascript:location.href='%s'\", 3000); </script></html>", reqPath)
		return
	}

	r.ParseMultipartForm(64 << 20) //64MB内存buffer
	var uplFail, upSucc int
	var rename string

	if r.MultipartForm == nil {
		fmt.Fprintf(w, "ERROR: MulitpartForm is nil\n")
		log.Println("r.MulitpartForm is nil")
		return
	}

	for _, fileHeader := range r.MultipartForm.File["uploadFiles"] {
		srcFile, err := fileHeader.Open()
		if err != nil {
			log.Printf("Open, %v", err)
			return
		}

		wd, _ := os.Getwd()
		dstPath := filepath.Join(wd, filepath.Base(fileHeader.Filename))
		originPath := dstPath
		upflag := "-上传"
		//循环检查上传文件是否和服务端文件重名，如果文件存在，则重命名上传文件,也就是“.扩展名” 前加 "upflagN"，加了后还重名就继续加
		for i := 1; ; i++ {
			_, err := os.Stat(dstPath)
			if err != nil {
				break
			}

			suffix := filepath.Ext(dstPath)
			withoutSuf := strings.TrimSuffix(dstPath, suffix)
			j := strings.LastIndex(withoutSuf, upflag)
			if j <= 0 {
				dstPath = withoutSuf + fmt.Sprintf("%s%d", upflag, i) + suffix
				continue
			}

			ind, err := strconv.Atoi(withoutSuf[j+len(upflag):])
			if err != nil {
				dstPath = withoutSuf + fmt.Sprintf("%s%d", upflag, i) + suffix
				continue
			}
			ind++
			i = ind
			dstPath = withoutSuf[:j] + fmt.Sprintf("%s%d", upflag, ind) + suffix
		}

		if originPath != dstPath {
			rename = rename + fmt.Sprintf("%-s&nbsp&nbsp<b>上传文件重名，上传文件被重命名为:</b>&nbsp&nbsp%-s </br>", originPath, dstPath)
		}

		dstFile, err := os.Create(dstPath) //创建上传文件
		if err != nil {
			log.Printf("os.Create, %v", err)
			srcFile.Close()
			return
		}

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			uplFail++
		} else {
			log.Printf("%s upload %q success\n", r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")], dstPath)
			upSucc++
		}
		srcFile.Close() //这里是循环，避免用defer *.Close()
		dstFile.Close()
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("Getwd(),err", err)
		return
	}
	pwd, err = filepath.Abs(pwd)
	if err != nil {
		log.Printf("Abs(),err", err)
		return
	}

	curDir := strings.TrimLeft(pwd, rootDir)
	if curDir == "" {
		curDir = "/"
	}
	fmt.Fprintf(w, "<h1>上传失败:%d, 成功:%d</h1> <p>%s</p> <script language='javascript' type='text/javascript'> setTimeout(\"javascript:location.href='%s'\", %d000); </script>",
		uplFail, upSucc, rename, curDir, 1+len(strings.Split(rename, "</br>")))
}

//获取目录/文件列表
func lsDir(dir string) uriBaseSize {
	var fi []os.FileInfo
	fi, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("ReadDir(),%v\n", err)
		return uriBaseSize{}
	}

	rdir, err := filepath.Rel(rootDir, dir)
	if err != nil {
		log.Printf("Rel(),%v\n", err)
		return uriBaseSize{}
	}
	back := filepath.FromSlash(filepath.Clean("/" + filepath.Dir(rdir)))

	var bs []baseSize
	var href string
	var base string
	for _, f := range fi {
		if f.IsDir() {
			base = f.Name() + "/"
		} else {
			base = f.Name()
		}
		href = url.PathEscape("/" + filepath.Join(rdir, base))
		bs = append(bs, baseSize{Href: href, Base: base, Size: f.Size()})
	}

	return uriBaseSize{Back: back, BS: bs}
}

// 将指定目录下的 "文件名-大小" 生成为html
func genHtml(dir string, w http.ResponseWriter) {
	ubs := lsDir(dir)
	tplt, e := template.New("temp0").Parse(htmlTemplate)
	if e != nil {
		log.Fatalf("Parse(): %v\n", e)
	}

	if e = tplt.Execute(w, ubs); e != nil {
		log.Fatalf("Execute() :%v\n", e)
	}
}

const htmlTemplate = `
<html>
<head>
    <title>WEB文件管理</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    <style type="text/css">
        table td { width:100%; border-bottom:dotted 1px red; }
        a{text-decoration:none}
    </style>
</head>

<body>
    <form enctype="multipart/form-data" action="/upload" method="post">
        <input type="file" multiple name="uploadFiles"/>
        <input type="submit" value="批量上传" />
    </form>

    <a style="font-weight:700; text-decoration: underline" href='/'>&#8634; 返回根目录</a></br>
    <a style="font-weight:700; text-decoration: underline" href='{{.Back}}'>&#8634; 返回上层目录</a>
    <table>
        <tr>
           <td style="font-weight:600">文件/目录名</td>
           <td style="font-weight:600">大小</h3></td>
        </tr>
    {{range $i,$bs := .BS}}
        <tr>
           <td><a href='{{$bs.Href}}'>&bull; {{$bs.Base}}</a></td>
           <td>{{$bs.Size}}</td>
        </tr>
    {{end}}
    </table>
</body>
</html>
`
