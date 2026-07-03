package postgres

import (
	"testing"
)

func FuzzMaskURL(f *testing.F) {
	for _, seed := range []string{
		"postgres://user:secret@localhost:5432/db",
		"postgres://user:p@ss@host:5432/db",
		"host=localhost user=u password=secret dbname=db",
		"password='quoted secret' host=localhost",
		"not a url",
		"",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 4096 {
			t.Skip()
		}
		out := MaskURL(input)
		_ = MaskURL(out)
	})
}
