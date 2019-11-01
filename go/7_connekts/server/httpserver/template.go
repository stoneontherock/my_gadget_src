package httpserver

import (
	"connekts/server/panicerr"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/url"
	"strconv"
	"strings"
	tt "text/template"
)

func init() {
	var err error
	errTmlp, _ = template.New("errTmpl").Parse(ERR_HTML)
	panicerr.Handle(err, "模板ERR_HTML解析错误.")

	listHostsTmpl, err = template.New("listHosts").Parse(LIST_HOSTS_HTML)
	panicerr.Handle(err, "模板LIST_ALIVE_HOSTS_HTML解析错误.")

	cmdOutTmpl, err = tt.New("cmd").Parse(CMD_HTML)
	panicerr.Handle(err, "模板CMD_HTML解析错误.")

	rPxyTmpl, err = template.New("rpxy").Parse(RPXY_HTML)
	panicerr.Handle(err, "模板RPXY_HTML解析错误.")

	//f := template.FuncMap{"filepathURLEscape": filepathURLEscape,"humanReadableSize":humanReadableSize}
	//listFileTmpl, err = template.New("listFS").Funcs(f).Parse(LIST_FILE_HTML)
	//panicerr.Handle(err, "模板LIST_FILE_HTML解析错误.")
}

var (
	errTmlp       *template.Template
	listHostsTmpl *template.Template
	cmdOutTmpl    *tt.Template
	rPxyTmpl      *template.Template
	listFileTmpl  *template.Template
)

func respJSAlert(c *gin.Context, code int, errStr string) {
	errTmlp.Execute(c.Writer, struct{ ERR string }{errStr})
}

func humanReadableSize(sz int32) string {
	switch {
	case sz < 1024:
		return strconv.Itoa(int(sz))
	case sz >= 1024 && sz < 1024*1024:
		return fmt.Sprintf("%.1fkB", float64(sz)/1024)
	case sz >= 1024*1024:
		return fmt.Sprintf("%.1fMB", float64(sz)/1024/1024)
	}

	return "0"
}

func filepathURLEscape(dir, base, mid string, fsize int32) string {
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	pth := dir + base
	if strings.HasSuffix(base, "/") {
		return fmt.Sprintf("/connekt/list_file?mid=%s&path=%s", mid, url.QueryEscape(pth))
	}

	return fmt.Sprintf("/connekt/file_up?mid=%s&path=%s&size=%d", mid, url.QueryEscape(pth), fsize)
}

const (
	ERR_HTML = `
<html>
<script>
	alert("{{.ERR}}");
</script>
</html>
`

	CMD_HTML = `
<!doctype html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>alives</title>
</head>
<body>
{{ $data := . -}}
<header>
    <h1>cmd</h1>
	<a href="/connekt/list_hosts">HOME</a>
</header>

<article>
    <form action="/connekt/cmd" method="POST">
        <input type="hidden"  name="mid" value="{{$data.MID}}" />
        <input type="text" name="timeout" value="60" />seconds timeout<br/>
        <textarea  rows="5" cols="100" name="cmd" placeholder='非shell环境执行需要用"..."分隔命令参数'></textarea> <br />
        <input type="submit" value="run" />
        <br />
    </form>

    {{- with $data.Stdout -}}
        <h3>stdout:</h3>
        <xmp>  {{- $data.Stdout -}} </xmp>
        <br />
    {{end}}

    {{- with $data.Stderr -}}
        <hr>
        <h3>stderr:</h3>
        <xmp> {{- $data.Stderr -}} </xmp>
    {{end}}
</article>
</body>
</html>
`

	RPXY_HTML = `
<!doctype html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>alives</title>
</head>
<body>
<header>
    <h1>rpxy</h1>
    <a href="/connekt/list_hosts">HOME</a>
</header>

<article>
    <form action="/connekt/rpxy" method="POST">
        <input type="hidden"  name="mid" value="{{.}}" />
        conn2连接数:<input type="text" name="num_of_conn2" value="1" placeholder='conn2连接数'/><br />
        port1:<input type="text" name="port1" placeholder='用户访问端端口号'/><br />
        addr3:<input type="text" name="addr3" placeholder='客户端需要反代的tcp地址'/><br />
        <input type="submit" value="rpxy" />
        <br />
    </form>
</article>
</body>
</html>
`

	//	LIST_FILE_HTML = `
	//<!doctype html>
	//<html lang="zh">
	//<head>
	//   <meta charset="UTF-8">
	//   <meta name="viewport"
	//         content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
	//   <meta http-equiv="X-UA-Compatible" content="ie=edge">
	//   <title>alives</title>
	//</head>
	//<body>
	//{{ $data := . -}}
	//<header>
	//   <h1>当前路径:{{- $data.Path -}}</h1>
	//   <hr>
	//</header>
	//
	//<article>
	//    {{- with $data.Fs -}}
	//    <table id="文件表格">
	//        <thead style="background-color: #EEEEFF;"><th style="text-align:left;">目录名/文件名</th><th style="text-align:right">大小</th></thead>
	//        <tbody>
	//        {{- range $index,$f := $data.Fs -}}
	//            <tr>
	//                <td class="col1"><a href="{{- filepathURLEscape $data.Path $f.Name $data.Mid $f.Size -}}"> &bull; {{- $f.Name -}} </a></td>
	//                <td class="col2">{{- humanReadableSize $f.Size -}}</td>
	//            </tr>
	//        {{end}}
	//        </tbody>
	//    </table>
	//    {{- else -}}
	//        <h3>{{- $data.Err -}}</h3>
	//    {{- end -}}
	//</article>
	//</body>
	//</html>
	//`

	LIST_HOSTS_HTML = `
<!doctype html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>alives</title>
</head>
<body>
{{ $data := . -}}
<header>
    <h1>All hosts</h1>
</header>

<article>
    <hr>
    <table id="文件表格">
        <thead style="background-color: #EEEEFF;"><th>index</th><th>MachineID</th><th>Hostname</th><th>OS</th><th>IP Addr</th><th>Ready</th><th>UpdateAt</th></thead>
        <tbody>
        {{- range $index,$rec := $data -}}
            <tr>
                <td>{{$index}}</td>
                <td>{{$rec.ID}}</td>
                <td>{{$rec.Hostname}}</td>
                <td>{{$rec.OS}}</td>
                <td>{{$rec.WanIP}}</td>
                <td>{{ if eq $rec.Pickup 1 }}
                        picking up
                    {{ else if eq $rec.Pickup 2 }}
                        ready
                    {{ else }}
                        free
                    {{ end }}
                </td>
                <td>{{$rec.UpdateAt}}</td>
                <td><form action="/connekt/del_host" method="POST">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="drop" />
                    </form>
                </td>
                {{ if lt $rec.Pickup 1 }}
                <td><form action="/connekt/change_pickup" method="POST">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="hidden"  name="pickup" value="1" />
                        <input type="submit" value="picking up" />
                    </form>
                </td>
                {{end}}
                {{ if eq $rec.Pickup 2 }}
                <td><form action="/connekt/cmd" method="POST">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="cmd" />
                    </form>
                </td>
                <td><form action="/connekt/rpxy" method="POST">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="rpxy" />
                    </form>
                </td>
                <!-- list_file 用不到
				<td><form action="/connekt/list_file" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="hidden"  name="path" value="/" />
                        <input type="submit" value="listFS" />
                    </form>
                </td>
                -->
                <td><form action="/connekt/filesystem" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="filesystem" />
                    </form>
                </td>
                {{end}}
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
