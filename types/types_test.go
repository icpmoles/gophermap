package types

import (
	"reflect"
	"testing"
)

func TestOptions(t *testing.T) {
	opts := Options{}

	options := []Option{
		WithAllowList([]string{"foo", "bar"}),
		WithUseExecutionTime(true),
		WithFrequency(Weekly),
	}

	for _, option := range options {
		option(&opts)
	}

	want := Options{
		AllowList:        []string{"foo", "bar"},
		UseExecutionTime: true,
		Frequency:        Weekly,
	}

	if !reflect.DeepEqual(opts, want) {
		t.Errorf("Options = %+v, want %+v", opts, want)
	}
}

func TestFrequencyFromString(t *testing.T) {
	var tests = []struct {
		input string
		want  SiteMapFrequency
	}{
		// already lowercase
		{"always", Always},
		{"hourly", Hourly},
		{"daily", Daily},
		{"weekly", Weekly},
		{"monthly", Monthly},
		{"yearly", Yearly},
		{"never", Never},
		// mixed with uppercase
		{"alwAys", Always},
		{"hOUrly", Hourly},
		{"daIly", Daily},
		{"WEEKlY", Weekly},
		{"mONthly", Monthly},
		{"YearlY", Yearly},
		{"NeveR", Never},

		// wrong, default to never
		{"Alw_ays", Never},
		{"_Always", Never},
		{"random", Never},
		{"weeklydaily", Never},
		{"mmonthly", Never},
		{" yearly", Never},
		{"neverever", Never},
	}

	for _, tt := range tests {
		// testname := fmt.Sprintf("%d,%d", tt.input, tt.want)
		t.Run(tt.input, func(t *testing.T) {
			opts := Options{}

			WithFrequencyFromString(tt.input)(&opts)

			if opts.Frequency != tt.want {
				t.Errorf("got %d, want %d", opts.Frequency, tt.want)
			}
		})
	}

}
