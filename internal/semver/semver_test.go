package semver

import (
	"testing"
)

func TestSatisfies(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		constraint string
		wantErr    bool
		errMsg     string
	}{
		// Bare major: "22" means ^22.0.0 (any 22.x.x)
		{name: "bare major match", version: "22.3.1", constraint: "22", wantErr: false},
		{name: "bare major mismatch", version: "20.1.0", constraint: "22", wantErr: true, errMsg: "does not satisfy"},
		{name: "bare major zero patch", version: "22.0.0", constraint: "22", wantErr: false},

		// Bare major.minor: "20.1" means ^20.1.0
		{name: "bare major.minor match", version: "20.1.5", constraint: "20.1", wantErr: false},
		{name: "bare major.minor mismatch minor", version: "20.0.5", constraint: "20.1", wantErr: true},
		{name: "bare major.minor mismatch major", version: "21.1.0", constraint: "20.1", wantErr: true},

		// Full version: "22.3.1" means ^22.3.1
		{name: "bare full match exact", version: "22.3.1", constraint: "22.3.1", wantErr: false},
		{name: "bare full match higher patch", version: "22.3.5", constraint: "22.3.1", wantErr: false},
		{name: "bare full mismatch lower patch", version: "22.3.0", constraint: "22.3.1", wantErr: true},

		// >= operator
		{name: ">= match", version: "22.3.1", constraint: ">= 20", wantErr: false},
		{name: ">= exact", version: "20.0.0", constraint: ">= 20", wantErr: false},
		{name: ">= mismatch", version: "18.5.0", constraint: ">= 20", wantErr: true, errMsg: "does not satisfy"},

		// ^ (caret) operator
		{name: "caret match", version: "18.5.0", constraint: "^18", wantErr: false},
		{name: "caret match minor", version: "18.5.3", constraint: "^18.0.0", wantErr: false},
		{name: "caret mismatch major", version: "19.0.0", constraint: "^18", wantErr: true},

		// ~ (tilde) operator
		{name: "tilde match", version: "20.1.5", constraint: "~20.1", wantErr: false},
		{name: "tilde mismatch minor", version: "20.2.0", constraint: "~20.1", wantErr: true},

		// Range
		{name: "range match", version: "19.0.0", constraint: ">= 18.0.0, < 20.0.0", wantErr: false},
		{name: "range lower bound", version: "18.0.0", constraint: ">= 18.0.0, < 20.0.0", wantErr: false},
		{name: "range upper mismatch", version: "20.0.0", constraint: ">= 18.0.0, < 20.0.0", wantErr: true},

		// v-prefix handling
		{name: "v-prefix version", version: "v22.3.1", constraint: "22", wantErr: false},
		{name: "v-prefix constraint", version: "22.3.1", constraint: "v22", wantErr: false},
		{name: "v-prefix both", version: "v22.3.1", constraint: "v22", wantErr: false},

		// Error cases
		{name: "empty version", version: "", constraint: "22", wantErr: true, errMsg: "invalid version"},
		{name: "empty constraint", version: "22.0.0", constraint: "", wantErr: true, errMsg: "invalid constraint"},
		{name: "garbage version", version: "abc", constraint: "22", wantErr: true, errMsg: "invalid version"},

		// Practical installer cases
		{name: "node major check", version: "22.12.0", constraint: "22", wantErr: false},
		{name: "python version check", version: "3.13.2", constraint: "3.13", wantErr: false},
		{name: "go version check", version: "1.23.4", constraint: "1.23", wantErr: false},
		{name: "go wrong minor", version: "1.22.4", constraint: "1.23", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Satisfies(tt.version, tt.constraint)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Satisfies(%q, %q) = nil, want error", tt.version, tt.constraint)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Satisfies(%q, %q) error = %q, want to contain %q", tt.version, tt.constraint, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Satisfies(%q, %q) = %v, want nil", tt.version, tt.constraint, err)
				}
			}
		})
	}
}

func TestNormalise(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"22", "22.0.0"},
		{"22.3", "22.3.0"},
		{"22.3.1", "22.3.1"},
		{"v22", "22.0.0"},
		{"v22.3", "22.3.0"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := normalise(tt.input)
			if err != nil {
				t.Fatalf("normalise(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("normalise(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
