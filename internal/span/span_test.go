package span

import "testing"

func TestTo(t *testing.T) {
	span := New(2, 5)
	other := New(7, 9)

	want := New(2, 9)
	got := span.To(other)
	if want != got {
		t.Errorf("%+v To(%+v) = %+v, want %+v\n", span, other, got, want)
	}
}
