package main

/*
var database_name string = "header_central_database"
var user_name string = "root"
var user_passwd string = "123123zx"
var database_ip string = "127.0.0.1:3306"
*/

var database_name string = "silicon_database"
var user_name string = "root"
var user_passwd string = "max@pwd**123.."
var database_ip string = "100.27.5.77:3306"

var dsn_str string = user_name + ":" + user_passwd + "@tcp(" + database_ip + ")/" + database_name