package dotenv_test

import (
	"testing"

	"github.com/profclems/go-dotenv"
)

func BenchmarkDotenv_Load(b *testing.B) {
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

func BenchmarkDotenv_instance(b *testing.B) {
	config := dotenv.New()
	config.SetConfigFile("fixtures/large.env")
	err := config.Load()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.Run("Get", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = config.Get("DB_USERNAME")
		}
	})

	b.ResetTimer()

	// benchmark Get for a key that does not exist
	b.Run("Get_NotExist", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = config.Get("DB_USERNAME_NOT_EXIST")
		}
	})

	b.ResetTimer()

	b.Run("Set", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			config.Set("APP_NAME", "My App")
		}
	})
}

func BenchmarkDotenv_global(b *testing.B) {
	dotenv.SetConfigFile("fixtures/large.env")
	err := dotenv.Load()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.Run("Get", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = dotenv.Get("DB_USERNAME")
		}
	})

	b.ResetTimer()
	// benchmark Get for a key that does not exist
	b.Run("Get_NotExist", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = dotenv.Get("DB_USERNAME_NOT_EXIST")
		}
	})

	b.ResetTimer()

	b.Run("Set", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			dotenv.Set("DB_USERNAME", "My App")
		}
	})
}
