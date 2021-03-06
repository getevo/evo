package generated

import (
	"time"
)

// SliceTest represents a single Slice conversion test.
type SliceTest struct {
	Into interface{}
	From []string
	Exp  interface{}
}

// SliceTests is a slice of all Slice conversion tests.
func NewSliceTests() (tests []SliceTest) {
	tests = []SliceTest{
		{new([]bool),
			[]string{"Yes", "FALSE"},
			[]bool{expBoolVal1, expBoolVal2}},
		{new([]*bool),
			[]string{"Yes", "FALSE"},
			[]*bool{&expBoolVal1, &expBoolVal2}},
		{new([]time.Duration),
			[]string{"10ns", "20µs", "30ms"},
			[]time.Duration{expDurationVal1, expDurationVal2, expDurationVal3}},
		{new([]*time.Duration),
			[]string{"10ns", "20µs", "30ms"},
			[]*time.Duration{&expDurationVal1, &expDurationVal2, &expDurationVal3}},
		{new([]float32),
			[]string{"1.2", "3.45", "6.78"},
			[]float32{expFloat32Val1, expFloat32Val2, expFloat32Val3}},
		{new([]*float32),
			[]string{"1.2", "3.45", "6.78"},
			[]*float32{&expFloat32Val1, &expFloat32Val2, &expFloat32Val3}},
		{new([]float64),
			[]string{"1.2", "3.45", "6.78"},
			[]float64{expFloat64Val1, expFloat64Val2, expFloat64Val3}},
		{new([]*float64),
			[]string{"1.2", "3.45", "6.78"},
			[]*float64{&expFloat64Val1, &expFloat64Val2, &expFloat64Val3}},
		{new([]int),
			[]string{"12", "34", "56"},
			[]int{expIntVal1, expIntVal2, expIntVal3}},
		{new([]*int),
			[]string{"12", "34", "56"},
			[]*int{&expIntVal1, &expIntVal2, &expIntVal3}},
		{new([]int16),
			[]string{"12", "34", "56"},
			[]int16{expInt16Val1, expInt16Val2, expInt16Val3}},
		{new([]*int16),
			[]string{"12", "34", "56"},
			[]*int16{&expInt16Val1, &expInt16Val2, &expInt16Val3}},
		{new([]int32),
			[]string{"12", "34", "56"},
			[]int32{expInt32Val1, expInt32Val2, expInt32Val3}},
		{new([]*int32),
			[]string{"12", "34", "56"},
			[]*int32{&expInt32Val1, &expInt32Val2, &expInt32Val3}},
		{new([]int64),
			[]string{"12", "34", "56"},
			[]int64{expInt64Val1, expInt64Val2, expInt64Val3}},
		{new([]*int64),
			[]string{"12", "34", "56"},
			[]*int64{&expInt64Val1, &expInt64Val2, &expInt64Val3}},
		{new([]int8),
			[]string{"12", "34", "56"},
			[]int8{expInt8Val1, expInt8Val2, expInt8Val3}},
		{new([]*int8),
			[]string{"12", "34", "56"},
			[]*int8{&expInt8Val1, &expInt8Val2, &expInt8Val3}},
		{new([]string),
			[]string{"k1", "K2", "03"},
			[]string{expStringVal1, expStringVal2, expStringVal3}},
		{new([]*string),
			[]string{"k1", "K2", "03"},
			[]*string{&expStringVal1, &expStringVal2, &expStringVal3}},
		{new([]time.Time),
			[]string{
				"2 Jan 2006 15:04:05 -0700 (UTC)",
				"Mon, 2 Jan 16:04:05 UTC 2006",
				"Mon, 02 Jan 2006 17:04:05 (UTC)"},
			[]time.Time{expTimeVal1, expTimeVal2, expTimeVal3}},
		{new([]*time.Time),
			[]string{
				"2 Jan 2006 15:04:05 -0700 (UTC)",
				"Mon, 2 Jan 16:04:05 UTC 2006",
				"Mon, 02 Jan 2006 17:04:05 (UTC)"},
			[]*time.Time{&expTimeVal1, &expTimeVal2, &expTimeVal3}},
		{new([]uint),
			[]string{"12", "34", "56"},
			[]uint{expUintVal1, expUintVal2, expUintVal3}},
		{new([]*uint),
			[]string{"12", "34", "56"},
			[]*uint{&expUintVal1, &expUintVal2, &expUintVal3}},
		{new([]uint16),
			[]string{"12", "34", "56"},
			[]uint16{expUint16Val1, expUint16Val2, expUint16Val3}},
		{new([]*uint16),
			[]string{"12", "34", "56"},
			[]*uint16{&expUint16Val1, &expUint16Val2, &expUint16Val3}},
		{new([]uint32),
			[]string{"12", "34", "56"},
			[]uint32{expUint32Val1, expUint32Val2, expUint32Val3}},
		{new([]*uint32),
			[]string{"12", "34", "56"},
			[]*uint32{&expUint32Val1, &expUint32Val2, &expUint32Val3}},
		{new([]uint64),
			[]string{"12", "34", "56"},
			[]uint64{expUint64Val1, expUint64Val2, expUint64Val3}},
		{new([]*uint64),
			[]string{"12", "34", "56"},
			[]*uint64{&expUint64Val1, &expUint64Val2, &expUint64Val3}},
		{new([]uint8),
			[]string{"12", "34", "56"},
			[]uint8{expUint8Val1, expUint8Val2, expUint8Val3}},
		{new([]*uint8),
			[]string{"12", "34", "56"},
			[]*uint8{&expUint8Val1, &expUint8Val2, &expUint8Val3}},
	}
	return
}
