package main

import (
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var dbFile = flag.String("sql", "", "sqlite数据库文件路径")
var port = flag.Int("port", 4096, "http监听端口")

type page struct {
	PageID         int
	AuthorUid      int
	AuthorUserName string
	ItemID         int
	CatId          int
	PageTitle      string
	PageContent    string
	SNumber        int
	AddTime        int
	PageComments   string
	IsDel          int
}

type catalog struct {
	CatId       int
	CatName     string
	ItemId      int
	SNumber     int
	Addtime     int
	ParentCatId int
	Level       int
}

var db *gorm.DB
var dirNameM = make(map[int]string)

func main() {
	flag.Parse()
	if *dbFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	http.HandleFunc("/", exportMD)
	err := http.ListenAndServe(":"+strconv.Itoa(*port), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func exportMD(wr http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		wr.Write([]byte(`
<html>
	<head>
		<title>showDoc导出MD</title>
	</head>
	<body>
	<form action="/" method="POST" enctype="x-www-form-urlencoded">					
		项目ID:<input type="number" name="project_id" required placeholder='例如：/web/#/7中的7'/>
		<input type="submit" value="提交"/>
	</form> 
	</body>
<html>
`))
		return
	}

	req.ParseForm()
	idstr := req.FormValue("project_id")
	prjID, err := strconv.Atoi(idstr)
	if err != nil || prjID == 0 {
		writeERR(wr, "无效的项目ID")
		return
	}

	db, err = gorm.Open("sqlite3", *dbFile)
	if err != nil {
		writeERR(wr, "sql Open:"+err.Error())
		return
	}
	db.SingularTable(true)
	defer db.Close()

	var project struct {
		ItemName string
	}
	db.Table(`item`).Select("item_name").Where(`item_id = ?`, prjID).Scan(&project)
	if project.ItemName == "" {
		writeERR(wr, "sql查询不到项目名")
		return
	}

	err = validateFilePath(project.ItemName)
	if err != nil {
		writeERR(wr, err.Error())
		return
	}

	var pgs []page
	err = db.Order("page_id asc").Find(&pgs, "item_id = ? AND is_del = 0", prjID).Error
	if err != nil {
		writeERR(wr, "sql:Find:"+err.Error())
		return
	}

	defer os.RemoveAll(project.ItemName)
	for _, p := range pgs {
		dir, ok := dirNameM[p.CatId]
		if !ok && p.CatId != 0 {
			dir, err = getCatDir(p.CatId)
			if err != nil {
				writeERR(wr, err.Error())
				return
			}
		}

		dir = project.ItemName + "/" + dir
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			writeERR(wr, "创建目录失败:"+err.Error())
			return
		}

		if err := validateFilePath(p.PageTitle); err != nil {
			writeERR(wr, err.Error())
			return
		}

		err = ioutil.WriteFile(dir+p.PageTitle+".md", []byte(p.PageContent), 0666)
		if err != nil {
			writeERR(wr, "写md文件失败:"+err.Error())
			return
		}
	}

	tgz := project.ItemName + ".tar.gz"
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("tar zcf '%s' '%s'", tgz, project.ItemName))
	err = cmd.Run()
	if err != nil {
		writeERR(wr, "压缩文件失败："+err.Error())
		return
	}
	defer os.RemoveAll(tgz)

	fp, err := os.Open(tgz)
	if err != nil {
		writeERR(wr, "打开压缩文件失败："+err.Error())
		return
	}
	defer fp.Close()

	wr.Header().Set("Content-Disposition", "attachment;filename="+tgz)
	_, err = io.Copy(wr, fp)
}

func writeERR(wr http.ResponseWriter, errStr string) {
	wr.WriteHeader(500)
	fmt.Fprintf(wr, `<html><h4>%s</h4> 三秒后跳转... <script language='javascript' type='text/javascript'> setTimeout("javascript:location.href='/'", 3000); </script></html>`, errStr)
}

func validateFilePath(pth string) error {
	if strings.Contains(pth, "/") || strings.Contains(pth, `\`) ||
		strings.Contains(pth, "<") || strings.Contains(pth, ">") ||
		strings.Contains(pth, "*") || strings.Contains(pth, "?") ||
		strings.Contains(pth, "|") || strings.Contains(pth, ":") ||
		strings.Contains(pth, `"`) || strings.Contains(pth, "'") {
		return fmt.Errorf("项目名和目录包含非法符号,项目名或目录名=%q", pth)
	}

	return nil
}

func getCatDir(catID int) (string, error) {
	dir := ""
	cid := catID
	for {
		c := catalog{}
		err := db.Where(`cat_id = ?`, cid).First(&c).Error
		if err != nil {
			return "", fmt.Errorf("getCatDir:查询cat_id失败: cat_id=" + strconv.Itoa(catID) + ":" + err.Error())
		}

		//fmt.Printf("***** %s\n",c.CatName)
		err = validateFilePath(c.CatName)
		if err != nil {
			return "", fmt.Errorf("项目名、目录名或标题(%q)包含非法符号", c.CatName)
		}

		dir = c.CatName + "/" + dir
		if c.ParentCatId == 0 {
			break
		}

		cid = c.ParentCatId
		time.Sleep(1e7)
	}

	return dir, nil
}
