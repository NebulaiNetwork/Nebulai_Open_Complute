package main

import (
	"fmt"
	"runtime"
	"path/filepath"
)

func DBG_LOG(v ...interface{}) {

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("Failed to get caller information")
		return
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		fmt.Println("Failed to get function information")
		return
	}
	
	path := file
    filename := filepath.Base(path)
	
	var outputStr string = "file[" + filename + "]\t| func[" + fn.Name() + "]\t| line[" + convertToString(line) + "]\t| log:"

	for _, val := range v {
		outputStr += convertToString(val)
	}

	fmt.Println(outputStr)
}

func DBG_ERR(v ...interface{}) {

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("Failed to get caller information")
		return
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		fmt.Println("Failed to get function information")
		return
	}
	
	path := file
    filename := filepath.Base(path)
	
	var outputStr string = "file[" + filename + "]\t| func[" + fn.Name() + "]\t| line[" + convertToString(line) + "]\t| log:"

	for _, val := range v {
		outputStr += convertToString(val)
	}

	fmt.Println(outputStr)
}


func splitStrAfterChar(str string, cutAfter rune) string {

	for i, char_ := range str {
		if char_ == cutAfter {
			return str[i+1:]
		}
	}

	return str
}

func splitStrByChar(str string, cutChar rune, v ...*string) {
	for _, val := range v {
		*val = str

		no_find_next_cut := true
		for i, char_ := range str {
			if char_ == cutChar {
				*val = str[:i]
				str = str[i+1:]
				no_find_next_cut = false
				break
			}
		}

		if no_find_next_cut {
			str = ""
		}
	}
}

func convertHEXStrToUint32(num string) uint32 {

	var ret uint32 = 0

	if len(num) > 8 {
		return ret
	}

	for _, data := range num {

		ret <<= 4

		var tmpNum uint8 = 0
		if data >= '0' && data <= '9' {
			tmpNum = uint8(data - '0')
		} else if data >= 'a' && data <= 'f' {
			tmpNum = uint8(data - 'a' + 10)
		} else if data >= 'A' && data <= 'F' {
			tmpNum = uint8(data - 'A' + 10)
		}

		ret += uint32(tmpNum)
	}
	return ret
}

func reverseStr(str string) string {
	var ret string = "0x"

	for i := len(str) - 1; i >= 0; i-- {
		ret += string(str[i])
	}

	return ret
}

func convertUint32ToHexString(num uint32) string {
	ret_str := ""

	for num != 0 {
		tmp_num := num & 0xF

		if tmp_num >= 0 && tmp_num <= 9 {
			ret_str += string(rune(tmp_num + '0'))
		} else {
			ret_str += string(rune(tmp_num + 'A' - 10))
		}

		num >>= 4
	}

	return reverseStr(ret_str)
}

func convertToString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func StrMsgToUint32Array(msg string) (int, []uint32) {
	ret_len := len(msg)/4 + 1
	ret := make([]uint32, ret_len)

	for i, _ := range ret {

		msg_len := len(msg)

		//fmt.Println(msg, "--", msg_len)

		if msg_len >= 4 {
			u1 := uint32(msg[0])
			u2 := uint32(msg[1])
			u3 := uint32(msg[2])
			u4 := uint32(msg[3])

			ret[i] = (u1 << 24) | (u2 << 16) | (u3 << 8) | u4
		} else {
			var u1, u2, u3 uint32 = 0, 0, 0

			switch msg_len {
			case 3:
				{
					u1 = uint32(msg[0])
					u2 = uint32(msg[1])
					u3 = uint32(msg[2])

					break
				}
			case 2:
				{
					u1 = uint32(msg[0])
					u2 = uint32(msg[1])

					break
				}
			case 1:
				{
					u1 = uint32(msg[0])

					break
				}
			}

			ret[i] = (u1 << 24) | (u2 << 16) | (u3 << 8)
		}
		if msg_len >= 4 {
			msg = msg[4:]
		}
	}

	return ret_len, ret
}

func RevertUint32ToStr(num_array []uint32) string {

	ret := ""

	for _, val := range num_array {
		u1 := uint8(val >> 24)
		u2 := uint8((val >> 16) & 0xFF)
		u3 := uint8((val >> 8) & 0xFF)
		u4 := uint8(val & 0xFF)

		ret += string(u1) + string(u2) + string(u3) + string(u4)
	}

	return ret
}

func convertIntStrToInt(num string) int {

	var ret int = 0
	
	for _, data := range num {
		ret *= 10
		ret += int(int(data - '0'))
	}
	return ret
}

