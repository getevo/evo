package utils

import (
	"math/rand"
	"strconv"
)

const (
	B           uint64 = 1
	KB                 = B << 10
	MB                 = KB << 10
	GB                 = MB << 10
	TB                 = GB << 10
	PB                 = TB << 10
	EB                 = PB << 10
	_MaxInt8_          = 1<<7 - 1
	_MinInt8_          = -1 << 7
	_MaxInt16_         = 1<<15 - 1
	_MinInt16_         = -1 << 15
	_MaxInt32_         = 1<<31 - 1
	_MinInt32_         = -1 << 31
	_MaxInt64_         = 1<<63 - 1
	_MinInt64_         = -1 << 63
	_MaxUint8_         = 1<<8 - 1
	_MaxUint16_        = 1<<16 - 1
	_MaxUint32_        = 1<<32 - 1
	_MaxUint64_        = 1<<64 - 1
)

//RandomBetween Create random number between two ranges
func RandomBetween(min, max int) int {
	return rand.Intn(max-min) + min
}

func ParseSafeInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func ParseSafeInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func ParseSafeFloat(s string) float64 {
	i, _ := strconv.ParseFloat(s, 10)
	return i
}

func MaxInt(args ...int) int {
	x := _MinInt32_
	for _, arg := range args {
		if arg > x {
			x = arg
		}
	}
	return x
}

func MaxUInt(args ...uint32) uint32 {
	var x uint32
	for _, arg := range args {
		if arg > x {
			x = arg
		}
	}
	return x
}

func MaxInt64(args ...int64) int64 {
	var x int64
	x = _MinInt64_
	for _, arg := range args {
		if arg > x {
			x = arg
		}
	}
	return x
}

func MinInt(args ...int) int {
	x := _MaxInt32_
	for _, arg := range args {
		if arg < x {
			x = arg
		}
	}
	return x
}

func MinUInt(args ...uint) uint {
	var x uint
	x = _MaxUint32_
	for _, arg := range args {
		if arg < x {
			x = arg
		}
	}
	return x
}

func MinInt64(args ...int64) int64 {
	var x int64
	x = _MaxInt64_
	for _, arg := range args {
		if arg < x {
			x = arg
		}
	}
	return x
}

func MinUInt64(args ...uint64) uint64 {
	var x uint64
	x = _MaxUint64_
	for _, arg := range args {
		if arg < x {
			x = arg
		}
	}
	return x
}

func AvgInt(args ...int) float64 {
	return float64(SumInt(args...)) / float64(len(args))
}

func AvgInt64(args ...int64) float64 {
	return float64(SumInt64(args...)) / float64(len(args))
}

func AvgUInt(args ...uint) float64 {
	return float64(SumUInt(args...)) / float64(len(args))
}

func AvgUInt64(args ...uint64) float64 {
	return float64(SumUInt64(args...)) / float64(len(args))
}

func SumInt(args ...int) int {
	var sum int
	for _, arg := range args {
		sum += arg
	}
	return sum
}

func SumInt64(args ...int64) int64 {
	var sum int64
	for _, arg := range args {
		sum += arg
	}
	return sum
}

//SumUInt Sum of uint32
func SumUInt(args ...uint) uint {
	var sum uint
	for _, arg := range args {
		sum += arg
	}
	return sum
}

func SumUInt64(args ...uint64) uint64 {
	var sum uint64
	for _, arg := range args {
		sum += arg
	}
	return sum
}
