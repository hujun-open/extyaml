package extyaml_test

import (
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/hujun-open/extyaml"
)

func TestMarshalDefault(t *testing.T) {
	extyaml.RegisterExt[time.Time](timeToStr, timeFromStr)
	type testCase struct {
		Input      ExampleStruct
		Default    ExampleStruct
		ExpectedRS string
	}

	doTest := func(c testCase) error {
		buf, err := extyaml.MarshalExtDefault(c.Input, c.Default)
		if err != nil {
			return err
		}
		t.Log(string(buf))
		if string(buf) != c.ExpectedRS {
			t.Log(c.ExpectedRS)
			return fmt.Errorf("expected result is different ")
		}
		return nil
	}
	caseList := []testCase{
		{
			Input: ExampleStruct{
				SubStruct: SubStruct{
					SubTimeSlice: []*time.Time{},
				},
				NonAnySub: SubStruct{
					SubTimeSlice: []*time.Time{},
				},
				AddrScalar: netip.MustParseAddr("1.1.1.1"),
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
				SlicePtrSub: []*SubStruct{
					&SubStruct{
						SubTimeSlice: []*time.Time{},
					},
				},
				MapSub: map[string]*SubStruct{"str1": &SubStruct{
					SubTimeSlice: []*time.Time{},
				}},
			},
			Default: ExampleStruct{
				AddrScalar: netip.MustParseAddr("1.1.1.1"),
				TimeScalar: time.Date(2022, 12, 1, 1, 2, 3, 0, time.UTC),
				TimeSlice: []time.Time{
					time.Date(2001, 02, 1, 1, 2, 3, 0, time.UTC),
					time.Date(2001, 03, 1, 1, 2, 3, 0, time.UTC),
					time.Date(2001, 04, 1, 1, 2, 3, 0, time.UTC),
				},
			},
			ExpectedRS: `substruct: {}
nonanysub: {}
strscalar: tom
timearray:
    - Fri, 01 Jan 2010 01:02:03 UTC
    - Wed, 01 Dec 2010 01:02:03 UTC
sliceptrsub:
    - subtimeslice: []
mapsub:
    str1:
        subtimeslice: []
`,
		},
	}
	for i, c := range caseList {
		t.Logf("test case %d", i)
		err := doTest(c)
		if err != nil {
			t.Fatalf("case %d failed, %v", i, err)
		}
	}

}
