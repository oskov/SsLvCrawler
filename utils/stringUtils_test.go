package utils

import "testing"

func Test_filterChars(t *testing.T) {
	type args struct {
		s       string
		pattern string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name:       "Money 40,000 $",
			args:       args{"40,000 $", "[^0-9]"},
			wantResult: "40000",
		},
		{
			name:       "Money 277 770 € ",
			args:       args{"277 770 € ", "[^0-9]"},
			wantResult: "277770",
		},
		{
			name:       "remove aaa",
			args:       args{"abaabbbbbaaabaaa", "(aaa)"},
			wantResult: "abaabbbbbb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := FilterChars(tt.args.s, tt.args.pattern); gotResult != tt.wantResult {
				t.Errorf("filterChars() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
