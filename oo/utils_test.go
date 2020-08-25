package oo

import (
	"testing"
)

func TestJsonUnmarshalValidate(t *testing.T) {
	{
		var (
			val struct{}
			buf = []byte(``)
		)
		err := JsonUnmarshalValidate(buf, &val)

		t.Log(err)
	}

	{
		var (
			val struct{}
			buf = []byte(`{"name":"test1"}`)
		)
		err := JsonUnmarshalValidate(buf, &val)

		t.Log(err)
	}

	{
		var (
			val struct {
				Name string `json:"name" validate:"gt=4"`
			}
			buf = []byte(``)
		)
		err := JsonUnmarshalValidate(buf, &val)

		t.Log(err)
	}

	{
		var (
			val struct {
				Name string `json:"name" validate:"gt=4"`
			}
			buf = []byte(`{"name":"test1"}`)
		)
		err := JsonUnmarshalValidate(buf, &val)

		t.Log(err)
	}
}

func TestStringsUniq(t *testing.T) {
	ss := []string{
		"aa", "bb", "cc", "bb", "dd", "ee",
	}
	ss = StringsUniq(ss, []string{"dd"})
	t.Log(ss)
}
