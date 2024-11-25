package extyaml_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/hujun-open/extyaml"
)

func compareStrs(str1, str2 string) error {
	if len(str1) != len(str2) {
		return fmt.Errorf("1st string is %d, while 2nd string is %d", len(str1), len(str2))
	}
	cmpstr := ""
	for i := 0; i < len(str1); i++ {
		if str1[i] != str2[i] {
			cmpstr += fmt.Sprintf("%d char is different, %c vs %c\n", i, str1[i], str2[i])
		}
	}
	if cmpstr != "" {
		return fmt.Errorf("diff: %v", cmpstr)
	}
	return nil
}

type Substract2 struct {
	At time.Time
}

type example2 struct {
	TimeScalar time.Time
	noExported int
	IsBool     bool
	Sub1       struct {
		Sub1int int
		noSub1  string
	}
	Sub2 Substract2
	Substract2
	IntSlice  []int
	TimeSlice []time.Time
	TimeArray [2]time.Time
}

func TestMarshalDefault(t *testing.T) {
	extyaml.RegisterExt[time.Time](timeToStr, timeFromStr)
	type testCase struct {
		Input      example2
		Default    example2
		ExpectedRS string
	}

	doTest := func(c testCase) error {
		buf, err := extyaml.MarshalExtDefault(c.Input, c.Default)
		if err != nil {
			return err
		}
		if string(buf) != c.ExpectedRS {
			t.Log(c.ExpectedRS)
			fmt.Println(compareStrs(c.ExpectedRS, string(buf)))
			return fmt.Errorf("expected marshal result %v is different from actual running result %v ", c.ExpectedRS, string(buf))
		}
		t.Log(string(buf))
		newval := c.Default
		err = extyaml.UnmarshalExt(buf, &newval)
		if err != nil {
			return err
		}
		inputBuf, err := extyaml.MarshalExt(c.Input)
		if err != nil {
			return fmt.Errorf("failed to marhsal input, %w", err)
		}
		newBuf, err := extyaml.MarshalExt(newval)
		if err != nil {
			return fmt.Errorf("failed to marhsal unmarshaled variable, %w", err)
		}
		if !bytes.Equal(inputBuf, newBuf) {
			return fmt.Errorf("unmarshaled variable is not equal to input")
		}

		return nil
	}
	caseList := []testCase{
		{
			Input: example2{
				TimeScalar: time.Date(2022, 12, 1, 1, 2, 3, 0, time.UTC),
				IsBool:     true,
				Sub1: struct {
					Sub1int int
					noSub1  string
				}{Sub1int: 10, noSub1: "xixi"},
				TimeSlice: []time.Time{
					time.Date(2020, 12, 1, 1, 2, 3, 0, time.UTC),
					time.Date(2021, 12, 1, 1, 2, 3, 0, time.UTC),
				},
				TimeArray: [2]time.Time{
					time.Date(2020, 12, 1, 1, 2, 3, 0, time.UTC),
					time.Date(2021, 12, 1, 1, 2, 3, 0, time.UTC),
				},
				Sub2: Substract2{
					At: time.Date(2033, 12, 1, 1, 2, 3, 0, time.UTC),
				},
			},
			Default: example2{
				TimeScalar: time.Date(2022, 12, 1, 1, 2, 3, 0, time.UTC),
				IsBool:     true,
				TimeSlice: []time.Time{
					time.Date(2020, 12, 1, 1, 2, 3, 0, time.UTC),
					time.Date(2021, 12, 1, 1, 2, 3, 0, time.UTC),
				},
				TimeArray: [2]time.Time{
					time.Date(2020, 12, 1, 1, 2, 3, 0, time.UTC),
					time.Date(2021, 12, 1, 1, 2, 3, 0, time.UTC),
				},
				Sub2: Substract2{
					At: time.Date(2033, 12, 1, 1, 2, 3, 0, time.UTC),
				},
			},
			ExpectedRS: `sub1:
    sub1int: 10
`,
		},
		//testing bool both true
		{
			Input: example2{
				IsBool: true,
			},
			Default: example2{
				IsBool: true,
			},
			ExpectedRS: `{}
`,
		},
		//testing bool case 1
		{
			Input: example2{
				IsBool: true,
			},
			Default: example2{
				IsBool: false,
			},
			ExpectedRS: `isbool: true
`,
		},
		//testing bool case 2
		{
			Input: example2{
				IsBool: false,
			},
			Default: example2{
				IsBool: true,
			},
			ExpectedRS: `isbool: false
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
