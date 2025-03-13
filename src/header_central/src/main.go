package main

import (
	"time"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		DBG_LOG("should set central id")
		return 
	}
	arg := os.Args[1]
	value, err := strconv.Atoi(arg)
    if err != nil {
        DBG_LOG("The argument is not a valid integer.")
        return
    }
	this_central_id = value
	DBG_LOG("header Central[", this_central_id, "] start")

	DBG_LOG("header central start")

	Init_Timer()
	
	Init_Default_Json_Type()
	go Start_SQL_Server()
	time.Sleep(2 * time.Second) //wait mysql init
	
	node_manager.InitByDatabase()

	tcp_server_manager.Init_Server()
	
	go Start_Http_Route()
	
	for {
		time.Sleep(1 * time.Second)
	}
}
