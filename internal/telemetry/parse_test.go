package telemetry

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr bool
		want    Report // checked only when wantErr is false
	}{
		{
			name: "valid line parses all fields",
			line: "v=1 id=web-abc seq=42 conns=7 rps=12.5",
			want: Report{Version: 1, ID: "web-abc", Seq: 42, Conns: 7, RPS: 12.5},
		},
		{
			name:    "missing id is an error",
			line:    "v=1 seq=1 conns=0 rps=0.0",
			wantErr: true,
		},
		{
			name:    "version 2 is rejected",
			line:    "v=2 id=web-abc seq=1 conns=0 rps=0.0",
			wantErr: true,
		},
		{
			name:    "token without equals is rejected",
			line:    "v=1 id=web-abc garbage seq=1",
			wantErr: true,
		},
		{
			name:    "duplicate key is rejected",
			line:    "v=1 id=web-abc id=web-xyz",
			wantErr: true,
		},
		{
			name: "unknown key is ignored",
			line: "v=1 id=web-abc future=hello seq=5 conns=1 rps=1.0",
			want: Report{Version: 1, ID: "web-abc", Seq: 5, Conns: 1, RPS: 1.0},
		},
		{
			name:    "non-numeric rps is rejected",
			line:    "v=1 id=web-abc seq=1 conns=0 rps=abc",
			wantErr: true,
		},
		{
			name:    "overlong line is rejected",
			line:    "v=1 id=" + strings.Repeat("x", 600),
			wantErr: true,
		},
		{
			name:    "negative conns is rejected",
			line:    "v=1 id=web-abc seq=1 conns=-5 rps=1.0",
			wantErr: true,
		},
		{
			name:    "NaN rps is rejected",
			line:    "v=1 id=web-abc seq=1 conns=0 rps=NaN",
			wantErr: true,
		},
		{
			name:    "infinite rps is rejected",
			line:    "v=1 id=web-abc seq=1 conns=0 rps=+Inf",
			wantErr: true,
		},
		        {
           		 name:    "missing version is an error",
           		 line:    "id=web-abc seq=1 conns=0 rps=0.0",
           		 wantErr: true,
       		 },
       		 {
           		 name:    "non-numeric conns is rejected",
           		 line:    "v=1 id=web-abc seq=1 conns=x rps=0.0",
           		 wantErr: true,
       		 },
		 {
           		 name:    "non-numeric seq is rejected",
           		 line:    "v=1 id=web-abc seq=x conns=0 rps=0.0",
           		 wantErr: true,
       		 },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.line)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) = %+v, want an error", tt.line, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.line, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.line, got, tt.want)
			}
		})
	}
}
