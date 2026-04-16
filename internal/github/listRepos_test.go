package github

import "testing"

func TestNormalizeVisibilityFilter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "empty defaults to all",
			input: "",
			want:  "all",
		},
		{
			name:  "all is preserved",
			input: "all",
			want:  "all",
		},
		{
			name:  "public is normalized",
			input: "PUBLIC",
			want:  "public",
		},
		{
			name:  "private trims whitespace",
			input: "  private  ",
			want:  "private",
		},
		{
			name:    "invalid value returns error",
			input:   "internal",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeVisibilityFilter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeVisibilityFilter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Fatalf("normalizeVisibilityFilter() = %q, want %q", got, tt.want)
			}
		})
	}
}
