package httpserver

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"line/server/panicerr"
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

	//listRProxiedTmpl, err = template.New("rpxy").Parse(LIST_RPROXIED_HTML)
	//panicerr.Handle(err, "模板LIST_RPXY_HTML解析错误.")
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
    <script>
        function BindEnter(obj)
        {
            //使用document.getElementById获取到按钮对象
            var button = document.getElementById('submitBtn');
            if(obj.keyCode == 13)
            {
                button.click();
                obj.returnValue = false;
            }
        }
    </script>
</head>
<body onkeydown="BindEnter(event)">
{{ $data := . -}}
<header>
	<a href="/line/list_hosts">返回主机管理界面</a>
</header>

<article>
    <form action="/line/cmd" method="GET">
        <input type="hidden"  name="mid" value="{{$data.MID}}" />
        <input type="text" name="timeout" value="15" />秒执行超时<br/>
        <input type="checkbox" name="inShell" value="true" checked/>在shell中执行<br/>
        <textarea  rows="5" cols="100" name="cmd" placeholder='输入命令，敲回车。linux支持多行'></textarea> <br />
        <input id="submitBtn" type="submit" value="执行" />
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
    <a href="/line/list_hosts">返回主机管理界面</a>
    <hr>
</header>

<article>
    <form action="/line/rpxy" method="GET">
        <input type="hidden"  name="mid" value="{{- $data.MID -}}" />
        line客户端和服务端预分配的连接数:<input type="text" name="num_of_conn2" value="1" placeholder='值大点，连接的速度会快一些'/><br />
        用户侧端口:<input type="text" name="port1" placeholder='用户访问端端口号'/><br />
        目标机地址:<input type="text" name="addr3" placeholder='客户端需要反代的tcp地址'/><br />
		标签:<input type="text" name="label" placeholder='给这条反代链路起个名'/><br />
        <input type="submit" value="执行反向代理" />
        <br />
    </form>
    <hr>
	{{- with $data.Labels -}}
	<h2>已经反向代理的主机列表</h2> <br/>
	<table id="rpxy table">
	    <thead style="background-color: #EEEEFF;"><th style="text-align:left;">标签:端口号</th><th style="text-align:right">操作</th></thead>
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

	LIST_HOSTS_HTML = `
<!doctype html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>alives</title>
	<style>
        #主机列表{
            border-collapse: collapse;
            border: 5px solid #EEEFFF;
			white-space: nowrap;
        }

        #主机列表 th,#主机列表 td{
            border: 5px solid #EEEFFF;
			white-space: nowrap;
        }

        .midSpan{
            display: inline-block;
            width:3em;
            height:1.2em;
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
        }

        .osInfoSpan{
            display: inline-block;
            width:8em;
            height: 1.2em;
            padding: 0px 0px;
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
        }

        .hoverSpan{
            display: none;
            position: absolute;
            background-color: black;
            color: white;
        }

        .briefSpan:hover~.hoverSpan{
                border: 1px solid grey;
                padding: 2px;
                display: block;
        }

        .opBtn {
            display: inline-block;
        }
	</style>
</head>
<body>
{{ $data := . -}}
<header>
	<div style="width=100%;text-align:right"><a href="/line/logout">退出</a></div>
</header>

<script>
	function pickup(mid) {
        let dur = prompt("勾起多少分钟后放下？",10);
        let req = new XMLHttpRequest();

        req.onreadystatechange=function(){
            if (req.readyState==4){ 
                if (req.status!=200){
            	    window.alert(req.responseText);
				}
				location="/line/list_hosts";
            }
        }
        req.open("GET","/line/change_pickup?pickup=1&timeout="+dur+"&mid="+mid,true);
        req.send();
    }
</script>

<article>
    <hr>
    <table id="主机列表">
        <thead style="background-color: #EEFFFF;"><th>机器ID</th><th>内核</th><th>OS信息</th><th>公网IP</th><th>心跳</th><th>状态</th><th>操作</th></thead>
        <tbody>
        {{- range $index,$rec := $data -}}
            <tr>
                <td><span class="midSpan briefSpan">{{$rec.ID}}</span><span class="hoverSpan">{{$rec.ID}}</span></td>
                <td>{{$rec.Kernel}}</td>
                <td><span class="osInfoSpan briefSpan">{{$rec.OsInfo}}</span><span class="hoverSpan">{{$rec.OsInfo}}</span></td>
                <td>{{$rec.WanIP}}</td>
                <td>{{$rec.Interval}}秒</td>
                <td>{{ if eq $rec.Pickup 1 }}
                        正在勾起...
                    {{ else if ge $rec.Pickup 2 }}
                        {{slice $rec.Timeout 8 10}}日{{slice $rec.Timeout 11 16}}释放
                    {{ else }}
                        未被勾住
                    {{ end }}
                </td>
		
                <td><form class="opBtn" action="/line/del_host" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="丢弃" />
                    </form>

                {{ if lt $rec.Pickup 1 }}
	                <button class="opBtn" onclick="pickup({{$rec.ID}})">勾住</button>
                {{end}}

                {{ if ge $rec.Pickup 2 }}
                <form class="opBtn" action="/line/cmd" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="命令" />
		</form>                
                <form class="opBtn" action="/line/rpxy" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="反代" />
                </form>
                
                <form class="opBtn" action="/line/filesystem" method="GET">
                        <input type="hidden"  name="mid" value="{{$rec.ID}}" />
                        <input type="submit" value="文件浏览" />
                </form>
                {{end}}
                </td>
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
<form action="/line/login" method="POST">
	用户名<input type="text"  name="user" required /><br/>
	密  码<input type="password"  name="pv" required /><br/>
    <input type="submit" value="登录" />
</form>
</body>
</html>
`
)
