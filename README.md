[![CI](https://github.com/hujun-open/extyaml/actions/workflows/main.yaml/badge.svg)](https://github.com/hujun-open/extyaml/actions/workflows/main.yaml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hujun-open/extyaml)](https://pkg.go.dev/github.com/hujun-open/extyaml)
# overview
extyaml is a golang module that does YAML marshaling/unmarshalling, it builds on top of [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml), to address following use case:

YAML support is needed for certain type that not in your control, e.g. a struct in a 3rd party module, but you don't want to create an alias type to add YAML support since that require changing the existing code to use this new type. 


## Usage
Provides a `FromStr` and `ToStr` function for each to-be-support type, and register them using `RegisterExt` function in `init()`. After registration, use `MarshalExt` for marshalling and `UnmarshalExt` for unmarshalling. if the type is already supported by `gopkg.in/yaml.v3`, then the registered functions overrides `gopkg.in/yaml.v3` marshaling/unmarshalling behavior.

Note: If a type implements one of following interface, it will be automatically used without need of registration:

- encoding.TextMarshaler/encoding.TextUnmarshaler
- gopkg.in/yaml.v3: Marshaler/Unmarshaler

Following is an example using custom layout string for `time.Time`, this overrides `gopkg.in/yaml.v3` marshaling/unmarshalling support for `time.Time`

```
package main

import (
	"fmt"
	"time"

	"github.com/hujun-open/extyaml"
)

type ExamplStruct struct {
	StrScalar   string
	TimePointer *time.Time
	TimeScalar  time.Time
	TimeArray   [2]time.Time
	TimeSlice   []time.Time
	TimeMap     map[time.Time]*time.Time
}

const myTimeLayout = "2006-Jan-02,Mon,15:04:05 MST" //the custom layout

func timeFromStr(s string) (any, error) {
	return time.Parse(myTimeLayout, s)
}

func timeToStr(in any) (string, error) {
	return in.(time.Time).Format(myTimeLayout), nil
}

func init() {
	//register time.Time, should be called before marshalling/unmarshalling
	extyaml.RegisterExt[time.Time](timeToStr, timeFromStr)
}

func main() {
	s := &ExamplStruct{
		StrScalar:  "example",
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
	}
	t := time.Date(1111, 12, 1, 1, 2, 3, 0, time.UTC)
	s.TimePointer = &t
	t = time.Date(2088, 02, 1, 1, 2, 3, 0, time.UTC)
	s.TimeMap = map[time.Time]*time.Time{
		time.Date(2099, 02, 1, 1, 2, 3, 0, time.UTC): &t,
	}
	//marshalling
	buf, err := extyaml.MarshalExt(s)
	if err != nil {
		panic(err)
	}
	fmt.Printf("marshaling result:\n\n%v\n", string(buf))
	//unmarshalling
	news := new(ExamplStruct)
	err = extyaml.UnmarshalExt(buf, news)
	if err != nil {
		panic(err)
	}
	fmt.Printf("unmarshaling result:\n%+v\n\n", news)
	//partial unmarshalling, where YAML only contains part of struct fields
	buf = []byte(`
timescalar: 2022-Dec-01,Thu,01:02:03 UTC
timearray:
    - 2010-Jan-01,Fri,01:02:03 UTC
    - 2010-Dec-01,Wed,01:02:03 UTC`)
	news = &ExamplStruct{
		StrScalar:  "init",
		TimeScalar: time.Date(4444, 12, 1, 1, 2, 3, 0, time.UTC),
		TimeArray: [2]time.Time{
			time.Date(5555, 01, 1, 1, 2, 3, 0, time.UTC),
			time.Date(5555, 12, 1, 1, 2, 3, 0, time.UTC),
		},
	}
	err = extyaml.UnmarshalExt(buf, news)
	if err != nil {
		panic(err)
	}
	fmt.Printf("partial unmarshaling result:\n%+v\n", news)
}

```
and following the output:
```
marshaling result:

strscalar: example
timepointer: 2088-Feb-01,Sun,01:02:03 UTC
timescalar: 2022-Dec-01,Thu,01:02:03 UTC
timearray:
    - 2010-Jan-01,Fri,01:02:03 UTC
    - 2010-Dec-01,Wed,01:02:03 UTC
timeslice:
    - 2001-Feb-01,Thu,01:02:03 UTC
    - 2001-Mar-01,Thu,01:02:03 UTC
    - 2001-Apr-01,Sun,01:02:03 UTC
timemap:
    2099-Feb-01,Sun,01:02:03 UTC: 2088-Feb-01,Sun,01:02:03 UTC

unmarshaling result:
&{StrScalar:example TimePointer:2088-02-01 01:02:03 +0000 UTC TimeScalar:2022-12-01 01:02:03 +0000 UTC TimeArray:[2010-01-01 01:02:03 +0000 UTC 2010-12-01 01:02:03 +0000 UTC] TimeSlice:[2001-02-01 01:02:03 +0000 UTC 2001-03-01 01:02:03 +0000 UTC 2001-04-01 01:02:03 +0000 UTC] TimeMap:map[2099-02-01 01:02:03 +0000 UTC:2088-02-01 01:02:03 +0000 UTC]}       

partial unmarshaling result:
&{StrScalar:init TimePointer:0001-01-01 00:00:00 +0000 UTC TimeScalar:2022-12-01 01:02:03 +0000 UTC TimeArray:[2010-01-01 01:02:03 +0000 UTC 2010-12-01 01:02:03 +0000 UTC] TimeSlice:[] TimeMap:map[]}
```

## Included Types

This module also include support for following types:

- `net.IPNet`: format as supported by `net.ParseCIDR`
- `net.HardwareAddr`
    - marshaling: xx:xx:xx:xx:xx:xx
    - unmarshalling: xx:xx:xx:xx:xx:xx, xx-xx-xx-xx-xx-xx