package internal_test

import (
	"reflect"
	"testing"

	utils "github.com/campfhir/wsv/internal"
)

func TestSplitEscaped(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "a,b,c",
			want:  []string{"a", "b", "c"},
		},
		{
			input: "a\\,b,c",
			want:  []string{"a,b", "c"},
		},
		{
			input: "one\\,two,three\\,four,five",
			want:  []string{"one,two", "three,four", "five"},
		},
		{
			input: "escaped\\\\,normal",
			want:  []string{"escaped\\", "normal"},
		},
		{
			input: "trailing\\,comma\\,",
			want:  []string{"trailing,comma,"},
		},
		{
			input: "",
			want:  []string{""},
		},
		{
			input: "\\,leading",
			want:  []string{",leading"},
		},
		{
			input: "double\\\\\\,escape",
			want:  []string{"double\\,escape"},
		},
		{
			input: "Date of Birth,date-format:Mon\\, Jan 02\\, 2006",
			want:  []string{"Date of Birth", "date-format:Mon, Jan 02, 2006"},
		},
	}

	for _, tt := range tests {
		got := utils.SplitEscaped(tt.input)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("SplitEscaped(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
