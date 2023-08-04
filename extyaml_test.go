package extyaml_test

import (
	"fmt"
	"net/netip"
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
	NonAnySub        SubStruct
	AddrScalar       netip.Addr
	AddrPointer      *netip.Addr
	TimePointer      *time.Time
	StrScalar        string
	TimeScalar       time.Time
	TimeArray        [2]time.Time
	TimeArrayPointer *[3]time.Time
	TimeSlice        []time.Time
	TimeMap          map[time.Time]*time.Time
	ShouldSkipAddr   netip.Addr `skipyamlmarshal:""`
	MapSub           map[string]*SubStruct
}

func TestExtyaml(t *testing.T) {
	extyaml.RegisterExt[time.Time](timeToStr, timeFromStr)
	origS := &ExamplStruct{
		AddrScalar: netip.AddrFrom4([4]byte{6, 7, 8, 9}),
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
		ShouldSkipAddr: netip.AddrFrom4([4]byte{2, 3, 4, 5}),
		StrScalar:      "tom",
		MapSub: map[string]*SubStruct{"str1": &SubStruct{
			SubTimeSlice: []*time.Time{},
		}},
	}
	tt99 := time.Date(1112, 12, 1, 1, 2, 3, 0, time.UTC)
	origS.MapSub["str1"].SubTimeSlice = append(origS.MapSub["str1"].SubTimeSlice, &tt99)
	addr1 := netip.AddrFrom4([4]byte{3, 4, 5, 6})
	origS.AddrPointer = &addr1
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
	buf, err := extyaml.MarshalExt(*origS)
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
addrscalar: 6.7.8.9
addrpointer: 3.4.5.6
timepointer: Fri, 01 Dec 1111 01:02:03 UTC
strscalar: tom
timescalar: Thu, 01 Dec 2022 01:02:03 UTC
timearray:
    - Fri, 01 Jan 2010 01:02:03 UTC
    - Wed, 01 Dec 2010 01:02:03 UTC
timearraypointer: null
timeslice:
    - Thu, 01 Feb 2001 01:02:03 UTC
    - Thu, 01 Mar 2001 01:02:03 UTC
    - Sun, 01 Apr 2001 01:02:03 UTC
timemap:
    Sun, 01 Feb 2099 01:02:03 UTC: Sun, 01 Feb 2088 01:02:03 UTC
mapsub:
    str1:
        subtimeslice:
            - Sun, 01 Dec 1112 01:02:03 UTC
`
	t.Log("marshaled result:\n" + string(buf))
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
	t.Logf("unmarshalled result:\n%+v", *newS)
	if !deepEqual(newS, origS) {

		t.Fatal("unmarshaled result is different from expected value")

	}
	if newS.TimePointer != origPointer {
		t.Fatal("pointer changed after unmarshalling")
	}
	//test partial unmarshalling
	partS := &ExamplStruct{
		TimeScalar:     time.Date(9999, 02, 1, 1, 2, 3, 0, time.UTC),
		ShouldSkipAddr: netip.AddrFrom4([4]byte{2, 3, 4, 5}),
	}
	tt3 := new(time.Time)
	*tt3 = time.Date(9999, 03, 1, 1, 2, 3, 0, time.UTC)
	partS.TimePointer = tt3
	origPointer = tt3
	expectedS := origS
	*expectedS.TimePointer = *partS.TimePointer
	expectedS.TimeArrayPointer = &([3]time.Time{
		time.Date(3999, 02, 1, 1, 2, 3, 0, time.UTC),
		time.Date(3998, 02, 1, 1, 2, 3, 0, time.UTC),
		time.Date(3997, 02, 1, 1, 2, 3, 0, time.UTC),
	})
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
timearraypointer:
    - Mon, 01 Feb 3999 01:02:03 UTC
    - Sun, 01 Feb 3998 01:02:03 UTC
    - Sat, 01 Feb 3997 01:02:03 UTC
timeslice:
    - Thu, 01 Feb 2001 01:02:03 UTC
    - Thu, 01 Mar 2001 01:02:03 UTC
    - Sun, 01 Apr 2001 01:02:03 UTC
timemap:
    Sun, 01 Feb 2099 01:02:03 UTC: Sun, 01 Feb 2088 01:02:03 UTC
timescalar: Thu, 01 Dec 2022 01:02:03 UTC
addrscalar: 6.7.8.9
addrpointer: 3.4.5.6
mapsub:
    str1:
        subtimeslice:
            - Sun, 01 Dec 1112 01:02:03 UTC
`
	err = extyaml.UnmarshalExt([]byte(partBuf), partS)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("partial unmarshaled result:\n%+v", *partS)
	if !deepEqual(partS, expectedS) {
		t.Fatal("partial unmarshaled result is different from expected value")
	}
	if partS.ShouldSkipAddr.String() != netip.AddrFrom4([4]byte{2, 3, 4, 5}).String() {
		t.Fatalf("parsed skip field is not expcted, %v", partS.ShouldSkipAddr)
	}
	if partS.TimePointer != origPointer {
		t.Fatal("pointer changed after partial unmarshalling")
	}

}

func deepEqual(in, expect any) bool {
	typeIn := reflect.TypeOf(in)
	typeExpect := reflect.TypeOf(expect)
	valIn := reflect.ValueOf(in)
	valExpect := reflect.ValueOf(expect)
	// fmt.Println("in", in, typeIn, "expect", expect, typeExpect, "-"+typeIn.PkgPath()+"-")
	if typeIn != typeExpect {
		return false
	}
	if typeIn.Kind() == reflect.Pointer {
		//compare to reflect.Value is to compare its interface()
		if valIn.Interface() == valExpect.Interface() {
			return true
		}

		if valIn.Elem().IsZero() && valExpect.IsNil() {
			return true
		}
		valIn = valIn.Elem()
		valExpect = valExpect.Elem()
		typeIn = typeIn.Elem()
	}

	switch typeIn.Kind() {
	case reflect.Struct:
		if typeIn.PkgPath() != "github.com/hujun-open/extyaml_test" && typeIn.PkgPath() != "" {
			return fmt.Sprint(valIn.Interface()) == fmt.Sprint(valExpect.Interface())
		}
		for i := 0; i < typeIn.NumField(); i++ {
			if _, exists := typeIn.Field(i).Tag.Lookup(extyaml.SkipTag); exists {
				continue
			}
			if !deepEqual(valIn.Field(i).Interface(), valExpect.Field(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Array, reflect.Slice:
		if valIn.Len() != valExpect.Len() {
			return false
		}
		for i := 0; i < valIn.Len(); i++ {
			if !deepEqual(valIn.Index(i).Interface(), valExpect.Index(i).Interface()) {
				return false
			}
		}
		return true

	}
	// fmt.Println(3333333333, valIn.Interface(), valExpect.Interface())
	return reflect.DeepEqual(valIn.Interface(), valExpect.Interface())
}
