package control

import (
	"testing"
)

// sink prevents dead code elimination by the compiler.
var sink any

// BenchmarkGenerateRequestID measures request ID generation with crypto/rand.
func BenchmarkGenerateRequestID(b *testing.B) {
	p := &Protocol{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := p.generateRequestID()
		sink = result
	}
}

// BenchmarkNewProtocol measures Protocol instantiation.
func BenchmarkNewProtocol(b *testing.B) {
	tests := []struct {
		name string
		opts []ProtocolOption
	}{
		{
			name: "minimal",
			opts: nil,
		},
		{
			name: "with_timeout",
			opts: []ProtocolOption{WithInitTimeout(30000)},
		},
		{
			name: "with_hooks",
			opts: []ProtocolOption{
				WithHooks(map[HookEvent][]HookMatcher{
					HookEventPreToolUse:  {{Matcher: "Bash"}},
					HookEventPostToolUse: {{Matcher: "Read"}},
				}),
			},
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result := NewProtocol(nil, tc.opts...)
				sink = result
			}
		})
	}
}
