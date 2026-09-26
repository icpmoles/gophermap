package main

import "testing"

func TestPrint_usage(t *testing.T) {
	print_usage()
}

func TestCliStringListConversion(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{[]string{"a"}, "a"},
		{[]string{"a", "ab"}, "a,ab"},
		{[]string{"a", "b", "c,"}, "a,b,c,"},
	}

	for i, test := range tests {
		var allowExt cliStringList

		for _, st := range test.input {
			allowExt.Set(st)

		}
		test_string := allowExt.String()
		if test_string != test.want {
			t.Fatalf("%d-th conversion not successfull: received '%s' instead of '%s'", i, test_string, test.want)

		}
	}

}
