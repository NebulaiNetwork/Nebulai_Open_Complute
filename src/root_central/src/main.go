package main

import (
	"time"
)

// ----------------Global Parameter>


// ----------------Function>


func main() {

	Init_Timer()

	Init_Default_Json_Type()
	go Start_SQL_Server()
	time.Sleep(2 * time.Second) //wait mysql init
	
	header_central_manager.Start_Server()
	go Start_Http_Route()

	for {
		time.Sleep(1 * time.Second)
	}
}
