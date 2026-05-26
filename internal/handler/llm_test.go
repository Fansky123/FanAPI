package handler

import "testing"

func TestHasTopLevelNonEmptyMessages(t *testing.T) {
	tests := []struct {
		name string
		req  map[string]interface{}
		want bool
	}{
		{
			name: "missing messages",
			req:  map[string]interface{}{},
			want: false,
		},
		{
			name: "empty messages",
			req: map[string]interface{}{
				"messages": []interface{}{},
			},
			want: false,
		},
		{
			name: "non-empty messages",
			req: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{"role": "user", "content": "hello"},
				},
			},
			want: true,
		},
		{
			name: "messages wrong type",
			req: map[string]interface{}{
				"messages": "not-array",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasTopLevelNonEmptyMessages(tt.req); got != tt.want {
				t.Fatalf("hasTopLevelNonEmptyMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}
