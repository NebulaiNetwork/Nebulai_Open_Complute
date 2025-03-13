package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "math/rand"
)

const (
    GetCentralIp = iota
    GetAllHeaderCentralInfo
    GetRouteIp
    CheckHeaderUidExist
    CheckWorkerUidExist
    CheckApiKeyExist
    RegNewAccount
    RegNewHeader
	RegNewWorker
	QueryHeader
	QueryWorker
	QueryCentralIdByHeaderUid
	QueryCentralIdByWorkerUid
)

type mysql_op_data struct {
	innerChannel	chan interface{}
	operator		int
	payload			string
	ip				string
	port			string
	uid				string
	account			string
	password		string
	central_id		string	//header central id
}

var mysql_ch chan mysql_op_data

func Start_SQL_Server(){

	mysql_ch = make(chan mysql_op_data)

	dsn := dsn_str
	DBG_LOG("database cmd==>", dsn)
	
	db, err := sql.Open("mysql", dsn)
    if err != nil {
        panic(err)
    }
    defer db.Close()
	
	for{
		data := <-mysql_ch

		switch data.operator {
		case CheckApiKeyExist:
			CheckApiKeyExistOp(db, data)
		case QueryCentralIdByHeaderUid:
			QueryCentralIdByHeaderUidOp(db, data)
		case QueryCentralIdByWorkerUid:
			QueryCentralIdByWorkerUidOp(db, data)
		case RegNewAccount:
			RegNewAccountOp(db, data)
		case RegNewHeader:
			RegNewHeaderOp(db, data)
		case QueryHeader:
			QueryHeaderOp(db, data)
		case QueryWorker:
			QueryWorkerOp(db, data)
		case RegNewWorker:
			RegNewWorkerOp(db, data)
		case GetCentralIp:
			QueryCentralIpAddress(db, data.innerChannel)
		case GetAllHeaderCentralInfo:
			GetAllHeaderCentralInfoOp(db, data.innerChannel)
		case GetRouteIp:
			GetRouteIpOp(db, data.innerChannel)
		}
	}
}

func CheckApiKeyExistOp(db *sql.DB, data mysql_op_data){
	
    rows, err := db.Query("SELECT count(api_key) FROM api_info where api_key = \"" + data.uid + "\";")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
        var count int
        err := rows.Scan(&count)
        if err != nil {
            panic(err)
        }

		if count == 1{
			data.innerChannel <- true
		}else{
			data.innerChannel <- false
		}
		return 
    }
}

func QueryCentralIdByHeaderUidOp(db *sql.DB, data mysql_op_data){	
    rows, err := db.Query("select header_central_id from header_info t1 join account_info t2 on t1.account_id=t2.id where t2.account=\"" + data.account + "\" and t2.password=\"" + data.password + "\" and t1.uid=\"" + data.uid + "\";")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
        var header_central_id int
        err := rows.Scan(&header_central_id)
        if err != nil {
            panic(err)
        }
        
		DBG_LOG("uid[", data.uid, "] at header_central[", header_central_id, "]")
		innerChannel <- header_central_id
		return 
    }
	DBG_ERR("uid[", data.uid, "] account[", data.account, "] password[", data.password,"] no find")
	innerChannel <- -1
}

func QueryCentralIdByWorkerUidOp(db *sql.DB, data mysql_op_data){
    rows, err := db.Query("select header_central_id from worker_info t1 join account_info t2 on t1.account_id=t2.id where t2.account=\"" + data.account + "\" and t2.password=\"" + data.password + "\" and t1.uid=\"" + data.uid + "\";")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
        var header_central_id int
        err := rows.Scan(&header_central_id)
        if err != nil {
            panic(err)
        }
        
		DBG_LOG("uid[", data.uid, "] at header_central[", header_central_id, "]")
		ret_chan <- header_central_id
		return 
    }
	DBG_ERR("uid[", data.uid, "] account[", data.account, "] password[", data.password,"] no find")
	ret_chan <- -1

}

func QueryCentralIpAddress(db *sql.DB, ret_chan chan interface{}){
	DBG_LOG("try query Central Ip")
	
    rows, err := db.Query("SELECT ip_address, port FROM central_info where id = 1;")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
        var ipAddress string
        var port string
        err := rows.Scan(&ipAddress, &port)
        if err != nil {
            panic(err)
        }
		
		ret := ipAddress + ":" + port
		DBG_LOG("get Central IP:", ret)
		ret_chan <- ret
		return 
    }
	DBG_LOG("error database read nil")
	ret_chan <- ""
}

func GetAllHeaderCentralInfoOp(db *sql.DB, ret_chan chan interface{}){
	DBG_LOG("try query header Central Ip")

	var ret []Header_Central_Info
	
    rows, err := db.Query("SELECT id, ip_address, api_port, tcp_port FROM header_central_info;")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
    	var id_ int
        var ipAddress string
        var api_port_ string
        var tcp_port_ string
        err := rows.Scan(&id_, &ipAddress, &api_port_, &tcp_port_)
        if err != nil {
            panic(err)
        }

		ret = append(ret, Header_Central_Info{id:id_, ip:ipAddress, api_port:api_port_, tcp_port:tcp_port_})
    }

    ret_chan <- ret
}

func GetRouteIpOp(db *sql.DB, ret_chan chan interface{}) {
	DBG_LOG("try query Route Ip")	
    rows, err := db.Query("SELECT ip_address, port FROM central_info where id = 2;")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
        var ipAddress string
        var port string
        err := rows.Scan(&ipAddress, &port)
        if err != nil {
            panic(err)
        }
		
		ret := ipAddress + ":" + port
		DBG_LOG("get Route IP:", ret)
		ret_chan <- ret
		return 
    }
	DBG_LOG("error database read nil")
	ret_chan <- ""
	
}

func ApiKeyIsExist(api_key string) bool{
	ch := make(chan interface{})

	mysql_ch <- mysql_op_data{innerChannel:ch, operator:CheckApiKeyExist, uid:api_key}

	result := <-ch

	return result.(bool)
}


func RegNewAccountOp(db *sql.DB, user_data mysql_op_data) {

	rows, err := db.Query("select count(id) from account_info where account=\"" + user_data.account + "\";")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
		var count_ int
		err := rows.Scan(&count_)
		if err != nil {
			panic(err)
		}

		if count_ > 0{
			user_data.innerChannel <- EXIST_ACCOUNT
			return 
		}
	}

	sqlCmd := "insert into account_info (account, password) values(?, ?);"
	
    _, err = db.Exec(sqlCmd, user_data.account, user_data.password)
    if err != nil {
		user_data.innerChannel <- FAILED
        DBG_LOG("reg account err:", err)
    }

	user_data.innerChannel <- SUCC
}


func RegNewHeaderOp(db *sql.DB, user_data mysql_op_data) {

	user_data.uid = CreateNewUid(db)

    sqlCmd := "insert into header_info (uid, header_central_id, working, account_id) select ?, ?, 0, id from account_info where account=? and password=?;"

    result, err := db.Exec(sqlCmd, user_data.uid, user_data.central_id, user_data.account, user_data.password)
    if err != nil {
		user_data.innerChannel <- FAILED
        DBG_LOG("RegNewHeaderOp err:", err)
    }

	rowsAffected, err2 := result.RowsAffected()
    if err2 != nil {
        DBG_ERR("get affect row error:", err2)
    }

    if rowsAffected == 1{
    	user_data.innerChannel <- SUCC
    } else{
		user_data.innerChannel <- FAILED

		DBG_LOG("affect row[", rowsAffected, "]")
    }
}

func RegNewWorkerOp(db *sql.DB, user_data mysql_op_data) {

	user_data.uid = CreateNewUid(db)

    sqlCmd := "insert into worker_info (uid, header_central_id, working, account_id) select ?, ?, 0, id from account_info where account=? and password=?;"

    result, err := db.Exec(sqlCmd, user_data.uid, user_data.central_id, user_data.account, user_data.password)
    if err != nil {
		user_data.innerChannel <- FAILED
        DBG_LOG("RegNewWorkerOp err:", err)
    }

    rowsAffected, err2 := result.RowsAffected()
    if err2 != nil {
        DBG_ERR("get affect row error:", err2)
    }

    if rowsAffected == 1{
    	user_data.innerChannel <- SUCC
    } else{
		user_data.innerChannel <- FAILED

		DBG_LOG("affect row[", rowsAffected, "]")
    }
}

func QueryHeaderOp(db *sql.DB, user_data mysql_op_data) {

	var und USER_NODES_DATA

	rows, err := db.Query("select t2.uid, t2.header_central_id from account_info t1 join header_info t2 on t1.id = t2.account_id where t1.account = \"" + user_data.account + "\" and t1.password = \"" + user_data.password + "\";")
    if err != nil {
    	DBG_ERR("mysql read error:", err)
    	und.Code = MYSQL_ERROR

    	user_data.innerChannel <- und
    	return
    }
    defer rows.Close()
	
    for rows.Next() {
		var uid string
		var header_central_id int
		err := rows.Scan(&uid, &header_central_id)
		if err != nil {
			DBG_ERR("mysql read error:", err)
    		und.Code = MYSQL_ERROR

    		break
		}

		und.Node_Uid = append(und.Node_Uid, uid)
		und.Central_Id = append(und.Central_Id, header_central_id)
	}

	user_data.innerChannel <- und
}

func QueryWorkerOp(db *sql.DB, user_data mysql_op_data) {

	var und USER_NODES_DATA

	rows, err := db.Query("select t2.uid, t2.header_central_id from account_info t1 join worker_info t2 on t1.id = t2.account_id where t1.account = \"" + user_data.account + "\" and t1.password = \"" + user_data.password + "\";")
    if err != nil {
    	DBG_ERR("mysql read error:", err)
    	und.Code = MYSQL_ERROR

    	user_data.innerChannel <- und
    	return
    }
    defer rows.Close()
	
    for rows.Next() {
		var uid string
		var header_central_id int
		err := rows.Scan(&uid, &header_central_id)
		if err != nil {
			DBG_ERR("mysql read error:", err)
    		und.Code = MYSQL_ERROR

    		break
		}

		und.Node_Uid = append(und.Node_Uid, uid)
		und.Central_Id = append(und.Central_Id, header_central_id)
	}

	user_data.innerChannel <- und
}

func CreateNewUid(db *sql.DB) string {

	var new_uid string

	for true{
		
		randomInt := rand.Intn(0xFFFFFFFF)
		new_uid = convertUint32ToHexString(uint32(randomInt))[2:]

		DBG_LOG("try new uid[", new_uid, "]")

		sqlCmd	:= "select count(id) from header_info where uid=\"" + new_uid + "\";"
		sqlCmd2	:= "select count(id) from worker_info where uid=\"" + new_uid + "\";"

		rows, err := db.Query(sqlCmd)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		rows2, err2 := db.Query(sqlCmd2)
		if err2 != nil {
			panic(err2)
		}
		defer rows2.Close()

		var count1 int
		var count2 int
		
		for rows.Next() {
			err := rows.Scan(&count1)
			if err != nil {
				DBG_ERR("read mysql result error")
				break
			}

		}

	
		for rows2.Next() {
			err := rows2.Scan(&count2)
			if err != nil {
				DBG_ERR("read mysql result error")
				break
			}
		}
		
		if count1 == count2 && count1 == 0{
			return new_uid
		}else{
			DBG_LOG("new uid count1[", count1, "] count2[", count2, "]")
		}
	}
	
	return new_uid
}

func Mysql_Op_No_Ret(op_data mysql_op_data){
	mysql_ch <- op_data
}

func Mysql_Op_Ret(op_data mysql_op_data)interface{}{
	ch := make(chan interface{})

	op_data.innerChannel = ch

	mysql_ch <- op_data

	ret := <-ch
	
	return ret
}

