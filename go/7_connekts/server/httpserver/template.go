package httpserver

import (
	"line/server/panicerr"
	"github.com/gin-gonic/gin"
	"html/template"
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

	AddRPxyTmpl, err = template.New("rpxy").Parse(ADD_RPXY_HTML)
	panicerr.Handle(err, "模板ADD_RPXY_HTML解析错误.")

	listRProxiedTmpl, err = template.New("rpxy").Parse(LIST_RPROXIED_HTML)
	panicerr.Handle(err, "模板LIST_RPXY_HTML解析错误.")
}

var (
	errTmlp          *template.Template
	listHostsTmpl    *template.Template
	cmdOutTmpl       *tt.Template
	AddRPxyTmpl      *template.Template
	listRProxiedTmpl *template.Template
)

func respJSAlert(c *gin.Context, code int, errStr string) {
	c.Status(code)
	errTmlp.Execute(c.Writer, struct{ ERR string }{errStr})
}

const (
	ERR_HTML = `
<html>
<script>
	alert("{{.ERR}}");
	location.href = "./list_hosts"
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
	<a href="/line/list_hosts">HOME</a>
</header>

<article>
    <form action="/line/cmd" method="GET">
        <input type="hidden"  name="mid" value="{{$data.MID}}" />
        <input type="text" name="timeout" value="30" />seconds timeout<br/>
        <input type="checkbox" name="inShell" value="true" checked/>exec in shell<br/>
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

	ADD_RPXY_HTML = `
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
    <h1>rpxy</h1>
    <a href="/line/list_hosts">HOME</a>
    <hr>
</header>

<article>
    <form action="/line/rpxy" method="GET">
        <input type="hidden"  name="mid" value="{{- $data.MID -}}" />
        conn2连接数:<input type="text" name="num_of_conn2" value="1" placeholder='conn2连接数'/><br />
        port1:<input type="text" name="port1" placeholder='用户访问端端口号'/><br />
        addr3:<input type="text" name="addr3" placeholder='客户端需要反代的tcp地址'/><br />
		label:<input type="text" name="label" placeholder='标签'/><br />
        <input type="submit" value="rpxy" />
        <br />
    </form>
    <hr>
	{{- with $data.Labels -}}
	<h2>rproxied</h2> <br/>
	<table id="rpxy table">
	    <thead style="background-color: #EEEEFF;"><th style="text-align:left;">label</th><th style="text-align:right">del</th></thead>
	    <tbody>
	          {{- range $index,$lab := $data.Labels -}}
					<tr>
	              		 <td class="col1">{{- $lab -}}</td>
						 <td class="col2"> 
							<form action="/line/del_rproxied" method="GET">
                        		<input type="hidden"  name="mid" value="{{- $data.MID -}}" />
                        		<input type="hidden"  name="label" value="{{- $lab -}}" />
								<input type="submit" value="✖️" />
		                    </form>
						 </td>
					</tr>
	          {{- end -}}
	    </tbody>
	</table>
	{{- end -}}
</article>
</body>
</html>
`

	LIST_RPROXIED_HTML = `
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
	  <a href="/line/list_hosts">HOME</a>
      <br />
	  <h1>总数:{{- len $data -}}</h1>
	  <hr>
	</header>
	
	<article>
	   {{- with $data -}}
	   <table id="rpxy table">
	       <thead style="background-color: #EEEEFF;"><th style="text-align:left;">mid</th><th style="text-align:right">label</th></thead>
	       <tbody>
	          {{- range $mid,$labs := $data -}}
					 {{- range $index,$lab := $labs -}}
						<tr>
	               		 <td class="col1">{{- $mid -}}</td>
	              		 <td class="col2">{{- $lab -}}</td>
						 <td class="col3"> 
							<form action="/line/del_rproxied" method="GET">
                        		<input type="hidden"  name="mid" value="{{- $mid -}}" />
                        		<input type="hidden"  name="label" value="{{- $lab -}}" />
								<input type="submit" value="✖️" />
		                    </form>
						 </td>
						</tr>
					{{- end -}}
	          {{- end -}}
	       </tbody>
	   </table>
	   {{- else -}}
	       <h3>{{- $data.Err -}}</h3>
	   {{- end -}}
	</article>
	</body>
	</html>
	`

	LIST_HOSTS_HTML = `
<!doctype html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>alives</title>
    <script>
	    function parseUnixTime(){
	        let ups = document.getElementsByClassName("updateAt");
    	    for (i=0;i<ups.length;i++) {
        	    let d = new Date(parseInt(ups[i].innerText)*1000);
            	ups[i].innerHTML = d.getMonth()+"-"+d.getDate()+" "+d.getHours()+":"+d.getMinutes()+":"+d.getSeconds();
        	}
    	}
	</script>
</head>
<body onload="parseUnixTime()">
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
                    {{ else if ge $rec.Pickup 2 }}
                        ready
                    {{ else }}
                        free
                    {{ end }}
                </td>
                <td class="updateAt">{{$rec.UpdateAt}}</td>
                <td><form action="/line/del_host" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="drop" />
                    </form>
                </td>

                {{ if lt $rec.Pickup 1 }}
                <td><form action="/line/change_pickup" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="hidden"  name="pickup" value="1" />
                        <input type="submit" value="picking up" />
                    </form>
                </td>
                {{end}}

                {{ if ge $rec.Pickup 2 }}
                <td><form action="/line/cmd" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="cmd" />
                    </form>
                </td>
                <td><form action="/line/rpxy" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="rpxy" />
                    </form>
                </td>
                <td><form action="/line/filesystem" method="GET">
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
	LOGIN_HTML = `
<html>
<head>
 <meta charset="UTF-8">
</head>
<body>
<form action="/login" method="GET">
	<input type="text"  name="user" required />
	<input type="password"  name="pv" required />
    <input type="submit" value="登录" />
</form>
</body>
</html>
`
)
