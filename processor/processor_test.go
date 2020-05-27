package processor

import (
	log "github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestProcess(t *testing.T) {
	type args struct {
		original string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{}, ""},
	}

	log.SetLevel(log.DebugLevel)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := "http://192.168.1.4/retailer/{{RandInRange(0,4)}}/store/{{RandInRange(5,9)}}?ids={{RandInRange(10,14)}}&word={{RandInList(hi,stalker)}}"
			t.Logf("Original: %s", original)
			process, _ := Process(original)
			t.Logf("Processed %s", process)
		})
	}
}

func Test_getPlaceholders(t *testing.T) {
	type args struct {
		template string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"Nothing to replace", args{"i'm just a /path or -  something"}, []string{}},
		{"Happy path", args{"{{$1()}}-{{$2(1,2,3)}}/{{$3(something)}} {{}}"}, []string{"{{$1()}}", "{{$2(1,2,3)}}", "{{$3(something)}}", "{{}}"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPlaceholders(tt.args.template)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPlaceholders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_funcBuilder(t *testing.T) {
	type args struct {
		name   string
		params []string
	}
	tests := []struct {
		name       string
		args       args
		shouldFail bool
	}{
		{"Build non valid function", args{"FakeFunc", []string{}}, true},
		{"Build RandInRange no params", args{"RandInRange", []string{}}, true},
		{"Build RandInRange wrong param number", args{"RandInRange", []string{"1", "2", "3"}}, true},
		{"Build RandInRange wrong first param type", args{"RandInRange", []string{"nan", "2"}}, true},
		{"Build RandInRange wrong second param type", args{"RandInRange", []string{"1", "nan"}}, true},
		{"Build RandInRange", args{"RandInRange", []string{"1", "2"}}, false},
		{"Build RandInList wrong param number", args{"RandInList", []string{}}, true},
		{"Build RandInList", args{"RandInList", []string{"hi", "stalker"}}, false},
		{"Build RandInFile wrong param number", args{"RandInFile", []string{}}, true},
		{"Build RandInFile", args{"RandInFile", []string{"fake-file.fake"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := funcBuilder(tt.args.name, tt.args.params)

			if !tt.shouldFail {
				if got == nil {
					t.Errorf("funcBuilder() should have returned a function")
				}
			} else {
				if err == nil {
					t.Errorf("funcBuilder() should have fail")
				}
			}
		})
	}
}
