package service

import "testing"

func TestEmailValidator_Syntax(t *testing.T) {
	v := NewEmailValidator()
	cases := []struct {
		in   string
		want bool
	}{
		{"good@example.com", true},
		{"  spaced@example.com  ", true},
		{"User.Name+tag@Domain.IO", true},
		{"missing-at.com", false},
		{"@nodomain.com", false},
		{"trailing@", false},
		{"", false},
		{"two@@signs.com", false},
	}
	for _, c := range cases {
		got, ok := v.Syntax(c.in)
		if ok != c.want {
			t.Errorf("Syntax(%q): got ok=%v want %v (normalized=%q)", c.in, ok, c.want, got)
		}
	}
}

func TestEmailValidator_Disposable(t *testing.T) {
	v := NewEmailValidator()
	cases := []struct {
		in   string
		want bool
	}{
		{"a@mailinator.com", true},
		{"b@10minutemail.com", true},
		{"c@yopmail.com", true},
		{"d@gmail.com", false},
		{"e@example.com", false},
		{"f@MAILINATOR.COM", true},
	}
	for _, c := range cases {
		if got := v.IsDisposable(c.in); got != c.want {
			t.Errorf("IsDisposable(%q): got %v want %v", c.in, got, c.want)
		}
	}
}
