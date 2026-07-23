package reader

import "testing"

// Regression test: WithInput used to do o.options = options, so a second
// call silently discarded whatever an earlier call had set — contradicting
// the function's own doc comment, which shows calling it twice (once for
// format+options, once more with "" to add more options).
func TestWithInput_AppendsAcrossCalls(t *testing.T) {
	var o opts
	if err := o.apply(
		WithInput("", "sample_rate=22050"),
		WithInput("", "channels=1"),
	); err != nil {
		t.Fatalf("apply: %v", err)
	}

	want := []string{"sample_rate=22050", "channels=1"}
	if len(o.options) != len(want) {
		t.Fatalf("options = %v, want %v", o.options, want)
	}
	for i, v := range want {
		if o.options[i] != v {
			t.Fatalf("options[%d] = %q, want %q", i, o.options[i], v)
		}
	}
}

func TestWithInput_InvalidFormat(t *testing.T) {
	var o opts
	if err := o.apply(WithInput("not_a_real_format")); err == nil {
		t.Fatal("apply: expected error for invalid input format")
	}
}
