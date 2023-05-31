package extyaml_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/hujun-open/extyaml"
)

const myTimeLayout = time.RFC1123 //"Mon, 02 Jan 2006 15:04:05 MST"

func timeFromStr(s string) (any, error) {
	return time.Parse(myTimeLayout, s)
}

func timeToStr(in any) (string, error) {
	return in.(time.Time).Format(myTimeLayout), nil
}

type SubStruct struct {
	SubTimeSlice []*time.Time
}

type ExamplStruct struct {
	SubStruct
	NonAnySub   SubStruct
	TimePointer *time.Time
	StrScalar   string
	TimeScalar  time.Time
	TimeArray   [2]time.Time
	TimeSlice   []time.Time
	TimeMap     map[time.Time]*time.Time
}

func TestExtyaml(t *testing.T) {
	extyaml.RegisterExt[time.Time](timeToStr, timeFromStr)
	origS := &ExamplStruct{
		TimeScalar: time.Date(2022, 12, 1, 1, 2, 3, 0, time.UTC),
		TimeArray: [2]time.Time{
			time.Date(2010, 01, 1, 1, 2, 3, 0, time.UTC),
			time.Date(2010, 12, 1, 1, 2, 3, 0, time.UTC),
		},
		TimeSlice: []time.Time{
			time.Date(2001, 02, 1, 1, 2, 3, 0, time.UTC),
			time.Date(2001, 03, 1, 1, 2, 3, 0, time.UTC),
			time.Date(2001, 04, 1, 1, 2, 3, 0, time.UTC),
		},
		StrScalar: "tom",
	}
	tt := time.Date(1111, 12, 1, 1, 2, 3, 0, time.UTC)
	origS.TimePointer = &tt
	tt2 := time.Date(2088, 02, 1, 1, 2, 3, 0, time.UTC)
	origS.TimeMap = map[time.Time]*time.Time{
		time.Date(2099, 02, 1, 1, 2, 3, 0, time.UTC): &tt2,
	}
	for i := 0; i < 3; i++ {
		newt := new(time.Time)
		*newt = time.Date(3330+i, 02, 1, 1, 2, 3, 0, time.UTC)
		origS.SubStruct.SubTimeSlice = append(origS.SubStruct.SubTimeSlice, newt)
	}
	for i := 0; i < 3; i++ {
		newt := new(time.Time)
		*newt = time.Date(4440+i, 02, 1, 1, 2, 3, 0, time.UTC)
		origS.NonAnySub.SubTimeSlice = append(origS.NonAnySub.SubTimeSlice, newt)
	}
	buf, err := extyaml.MarshalExt(origS)
	if err != nil {
		t.Fatal(err)
	}
	expectResult := `substruct:
    subtimeslice:
        - Wed, 01 Feb 3330 01:02:03 UTC
        - Thu, 01 Feb 3331 01:02:03 UTC
        - Fri, 01 Feb 3332 01:02:03 UTC
nonanysub:
    subtimeslice:
        - Wed, 01 Feb 4440 01:02:03 UTC
        - Fri, 01 Feb 4441 01:02:03 UTC
        - Sat, 01 Feb 4442 01:02:03 UTC
timepointer: Fri, 01 Dec 1111 01:02:03 UTC
strscalar: tom
timescalar: Thu, 01 Dec 2022 01:02:03 UTC
timearray:
    - Fri, 01 Jan 2010 01:02:03 UTC
    - Wed, 01 Dec 2010 01:02:03 UTC
timeslice:
    - Thu, 01 Feb 2001 01:02:03 UTC
    - Thu, 01 Mar 2001 01:02:03 UTC
    - Sun, 01 Apr 2001 01:02:03 UTC
timemap:
    Sun, 01 Feb 2099 01:02:03 UTC: Sun, 01 Feb 2088 01:02:03 UTC
`
	t.Log("\n" + string(buf))
	if expectResult != string(buf) {
		t.Fatal("marshaling result is different from expcted result")
	}
	newS := new(ExamplStruct)
	origPointer := new(time.Time)
	newS.TimePointer = origPointer
	err = extyaml.UnmarshalExt(buf, newS)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(newS, origS) {
		t.Fatal("unmarshaled result is different from expected value")
	}
	if newS.TimePointer != origPointer {
		t.Fatal("pointer changed after unmarshalling")
	}
	//test partial unmarshalling
	partS := &ExamplStruct{
		TimeScalar: time.Date(9999, 02, 1, 1, 2, 3, 0, time.UTC),
	}
	tt3 := new(time.Time)
	*tt3 = time.Date(9999, 03, 1, 1, 2, 3, 0, time.UTC)
	partS.TimePointer = tt3
	origPointer = tt3
	expectedS := origS
	*expectedS.TimePointer = *partS.TimePointer
	partBuf := `substruct:
    subtimeslice:
        - Wed, 01 Feb 3330 01:02:03 UTC
        - Thu, 01 Feb 3331 01:02:03 UTC
        - Fri, 01 Feb 3332 01:02:03 UTC
nonanysub:
    subtimeslice:
        - Wed, 01 Feb 4440 01:02:03 UTC
        - Fri, 01 Feb 4441 01:02:03 UTC
        - Sat, 01 Feb 4442 01:02:03 UTC
strscalar: tom
timearray:
    - Fri, 01 Jan 2010 01:02:03 UTC
    - Wed, 01 Dec 2010 01:02:03 UTC
timeslice:
    - Thu, 01 Feb 2001 01:02:03 UTC
    - Thu, 01 Mar 2001 01:02:03 UTC
    - Sun, 01 Apr 2001 01:02:03 UTC
timemap:
    Sun, 01 Feb 2099 01:02:03 UTC: Sun, 01 Feb 2088 01:02:03 UTC
timescalar: Thu, 01 Dec 2022 01:02:03 UTC
`
	err = extyaml.UnmarshalExt([]byte(partBuf), partS)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(partS, expectedS) {
		t.Fatal("unmarshaled result is different from expected value")
	}
	if partS.TimePointer != origPointer {
		t.Fatal("pointer changed after partial unmarshalling")
	}

}
