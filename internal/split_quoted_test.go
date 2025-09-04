package internal_test

import (
	"reflect"
	"testing"

	"github.com/campfhir/wsv/internal"
)

func TestSplitQuoted(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "a,b,c",
			want:  []string{"a", "b", "c"},
		},
		{
			input: "'a,b',c",
			want:  []string{"a,b", "c"},
		},
		{
			input: "'one, two','three, four',five",
			want:  []string{"one, two", "three, four", "five"},
		},
		{
			input: "normal,'with \\'quote\\'',end",
			want:  []string{"normal", "with 'quote'", "end"},
		},
		{
			input: "'escaped \\\\', still inside',outside",
			want:  []string{"escaped \\, still inside", "outside"},
		},
		{
			input: "''",
			want:  []string{""}, // empty quoted string
		},
		{
			input: "a,'',b",
			want:  []string{"a", "", "b"}, // empty quoted field
		},
		{
			input: "",
			want:  []string{""}, // empty input
		},
		{
			input: "'unterminated",
			want:  []string{"unterminated"}, // quotes never closed
		},
	}

	for _, tt := range tests {
		got := internal.SplitQuoted(tt.input)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("SplitQuoted(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}

}
