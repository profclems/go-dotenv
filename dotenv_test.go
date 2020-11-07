package dotenv

import "testing"

func BenchmarkGetFromFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GetFromFile(DefaultConfigFile, "foo")
	}
}
