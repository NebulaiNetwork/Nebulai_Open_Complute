package main

import(
	"fmt"
	"time"
    "math/rand"
)

func GetTypeBitOfNum(num uint32, checkLen uint8, isSet bool) int {
	var ret int = 0

	if isSet {
		for ; checkLen != 0; checkLen-- {
			if (num & 1) == 1 {
				ret++
			}
			num >>= 1
		}
	} else {
		for ; checkLen != 0; checkLen-- {
			if (num & 1) == 0 {
				ret++
			}
			num >>= 1
		}
	}

	return ret
}

func Encode(originNum uint32) string {
	rand.Seed(time.Now().UnixNano())

	o1 := uint8(originNum & 0xFF)
	o2 := uint8((originNum >> 8) & 0xFF)
	o3 := uint8((originNum >> 16) & 0xFF)
	o4 := uint8((originNum >> 24) & 0xFF)

	var r1, r2, A, B, C, D, E, F uint8

	randomInt := rand.Intn(65536)

	r1 = uint8((randomInt >> 8) % 0xFF)
	r2 = uint8(randomInt % 0xFF)
	F = r2 ^ o3 ^ o4
	E = r1 ^ o2 ^ o3
	D = r2 ^ o1 ^ o2

	var A_B uint16 = 0
	tmpNum := 0

	if randomInt > 32500 {
		A_B = 1
		tmpNum = GetTypeBitOfNum(originNum, 32, true)
	} else {
		A_B = 0
		tmpNum = GetTypeBitOfNum(originNum, 32, false)
	}
	A_B |= uint16(tmpNum << 1)

	if rand.Intn(100) > 49 {
		A_B |= 1 << 6
		originNum >>= 1
	}
	A_B |= uint16((originNum & 1) << 7)

	randomInt = rand.Intn(100) & 0x3
	A_B |= uint16(randomInt << 8)

	originNum >>= randomInt
	A_B |= uint16((originNum & 1) << 10)

	randomInt = rand.Intn(65536)
	randomInt = randomInt & 0xF
	A_B |= uint16(randomInt << 11)

	originNum >>= randomInt
	A_B |= uint16((originNum & 1) << 15)

	A = uint8(A_B >> 8)
	B = uint8(A_B & 0xFF)
	C = A ^ B ^ r1 ^ o1

	var ret_num uint64

	ret_num = (uint64(r1) << 56) | (uint64(r2) << 48) | (uint64(A) << 40) | (uint64(B) << 32) | (uint64(C) << 24) | (uint64(D) << 16) | (uint64(E) << 8) | uint64(F)

	ret_str := fmt.Sprintf("%016x", ret_num)

	return ret_str
}
