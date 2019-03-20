package main

import (
	"log"
	"os"
	wm "webmasq"
)

func main() {
	wm.InitLog()
	wm.AddUser("zh", "uohz")
	log.Fatal(wm.Serve(os.Args[1], os.Args[2], os.Args[3], os.Args[4]))
}
