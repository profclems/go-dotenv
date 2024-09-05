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
		err := config.LoadConfig()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDotenv_Init_GetSet(b *testing.B) {
	config := dotenv.New()
	config.SetConfigFile("fixtures/large.env")
	err := config.LoadConfig()
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Get", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = config.GetString("APP_NAME")
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
	err := dotenv.LoadConfig()
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
