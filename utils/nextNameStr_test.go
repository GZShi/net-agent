package utils

import "testing"

func TestNextNameStr(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case1", args{"hello_1233"}, "hello_1234"},
		{"case2", args{"hello_001"}, "hello_002"},
		{"case2", args{"hello_"}, "hello_1"},
		{"case2", args{"hello"}, "hello_1"},
		{"case2", args{"hello_1_2_3"}, "hello_1_2_4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NextNameStr(tt.args.str); got != tt.want {
				t.Errorf("NextNameStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
