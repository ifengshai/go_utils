package main

import "github.com/ifengshai/go_utils"

func main() {
	//创建map
	var dist = make(map[string]string)
	dist["z"] = "a"

	go_utils.Printf(dist)

	a := 5
	go_utils.Printf(a)
}
