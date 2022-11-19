package shlex_test

import (
	"strings"
	"testing"

	"github.com/midbel/shlex"
)

var list = []struct {
	Input string
	Want  []string
}{
	{
		Input: `echo`,
		Want:  []string{"echo"},
	},
	{
		Input: `echo -e foo    bar`,
		Want:  []string{"echo", "-e", "foo", "bar"},
	},
	{
		Input: `echo foo bar`,
		Want:  []string{"echo", "foo", "bar"},
	},
	{
		Input: `echo "foo bar"`,
		Want:  []string{"echo", "foo bar"},
	},
	{
		Input: `echo 'foo bar'`,
		Want:  []string{"echo", "foo bar"},
	},
	{
		Input: `echo; echo | cat |& cut; echo && cut;`,
		Want:  []string{"echo", ";", "echo", "|", "cat", "|&", "cut", ";", "echo", "&&", "cut", ";"},
	},
	{
		Input: `echo ${var}`,
		Want:  []string{"echo", "${var}"},
	},
	{
		Input: `echo $var`,
		Want:  []string{"echo", "$var"},
	},
	{
		Input: `echo $(echo | cat | cut)`,
		Want:  []string{"echo", "$(echo | cat | cut)"},
	},
	{
		Input: `echo $(echo; echo $(cut))`,
		Want:  []string{"echo", "$(echo; echo $(cut))"},
	},
	{
		Input: `echo $((1+1))`,
		Want:  []string{"echo", "$((1+1))"},
	},
	{
		Input: `echo $((1+1*(2-1)))`,
		Want:  []string{"echo", "$((1+1*(2-1)))"},
	},
	{
		Input: `echo ${var#prefix}`,
		Want:  []string{"echo", "${var#prefix}"},
	},
	{
		Input: `echo ${var/from/to}`,
		Want:  []string{"echo", "${var/from/to}"},
	},
	{
		Input: `echo {1,2,3}`,
		Want:  []string{"echo", "{1,2,3}"},
	},
	{
		Input: `echo {1..10}`,
		Want:  []string{"echo", "{1..10}"},
	},
	{
		Input: `echo {}`,
		Want:  []string{"echo", "{}"},
	},
	{
		Input: `echo foobar # a comment`,
		Want:  []string{"echo", "foobar", "# a comment"},
	},
	{
		Input: `echo (echo foobar)`,
		Want:  []string{"echo", "(echo foobar)"},
	},
	{
		Input: `echo (echo foobar (echo foobar))`,
		Want:  []string{"echo", "(echo foobar (echo foobar))"},
	},
	{
		Input: `echo { echo foobar; }`,
		Want:  []string{"echo", "{ echo foobar; }"},
	},
}

func TestSplit(t *testing.T) {
	for _, in := range list {
		str, err := shlex.Split(strings.NewReader(in.Input))
		if err != nil {
			t.Errorf("%s: unexpected error! %s", in.Input, err)
			continue
		}
		if len(str) != len(in.Want) {
			t.Errorf("%s: length mismatched! got %d, want %d", in.Input, len(str), len(in.Want))
			t.Logf("got:  %q", str)
			t.Logf("want: %q", in.Want)
			continue
		}
		for i := range str {
			if str[i] != in.Want[i] {
				t.Errorf("word mismatched! got %s, want %s", str[i], in.Want[i])
			}
		}
	}
}
