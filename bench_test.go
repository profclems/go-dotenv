package dotenv_test

import (
	"testing"

	"github.com/profclems/go-dotenv"
)

func BenchmarkDotenv_LoadConfig(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		config := dotenv.New()
		config.SetConfigFile("fixtures/large.env")
		err := config.Load()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDotenv_Init_GetSet(b *testing.B) {
	config := dotenv.New()
	config.SetConfigFile("fixtures/large.env")
	err := config.Load()
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Get", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = config.Get("APP_NAME")
		}
	})

	b.Run("Set", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			config.Set("APP_NAME", "My App")
		}
	})
}

func BenchmarkDotenv_LoadConfig_GetSet(b *testing.B) {
	dotenv.SetConfigFile("fixtures/large.env")
	err := dotenv.Load()
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Get", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = dotenv.Get("APP_NAME")
		}
	})

	b.Run("Set", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			dotenv.Set("APP_NAME", "My App")
		}
	})
}
