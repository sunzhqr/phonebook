package normalizer_test

import (
	"testing"

	"github.com/sunzhqr/phonebook/pkg/normalizer"
)

func Test_NormalizePhone(t *testing.T) {
	e164, digits, ok := normalizer.NormalizePhone(" +7 (771) 123-45-67 ")
	if !ok || e164 != "+77711234567" || digits != "77711234567" {
		t.Fatalf("bad normalize: e164=%s digits=%s ok=%v", e164, digits, ok)
	}
	_, _, ok = normalizer.NormalizePhone("123")
	if ok {
		t.Fatalf("too short should be false")
	}
}
