package install

import "testing"

func TestHumanBytes(t *testing.T) {
	tests := map[int64]string{
		512:        "512 B",
		1024:       "1.0 KiB",
		1048576:    "1.0 MiB",
		1073741824: "1.0 GiB",
	}

	for input, want := range tests {
		if got := humanBytes(input); got != want {
			t.Fatalf("humanBytes(%d) = %q, want %q", input, got, want)
		}
	}
}
