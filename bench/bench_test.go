package bench

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/profclems/go-dotenv"
	"github.com/spf13/viper"
)

func BenchmarkDotenv_Init(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		config := dotenv.Init("../fixtures/normal.env")
		err := config.LoadConfig()
		if err != nil {
			b.Fatal(err)
		}

		_ = config.GetString("S3_BUCKET")
		_ = config.GetString("SECRET_KEY")
		_ = config.GetInt("PRIORITY_LEVEL")
	}
}

func BenchmarkDotenv_LoadConfig(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dotenv.SetConfigFile("../fixtures/normal.env")
		err := dotenv.LoadConfig()
		if err != nil {
			b.Fatal(err)
		}

		_ = dotenv.GetString("S3_BUCKET")
		_ = dotenv.GetString("SECRET_KEY")
		_ = dotenv.GetInt("PRIORITY_LEVEL")
	}
}

func BenchmarkJohoGodotenv(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := godotenv.Load("../fixtures/normal.env")
		if err != nil {
			b.Fatal(err)
		}

		_ = os.Getenv("S3_BUCKET")
		_ = os.Getenv("SECRET_KEY")
		_ = os.Getenv("PRIORITY_LEVEL")
	}
}

func BenchmarkViper_New(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		viper := viper.New()
		viper.SetConfigFile("../fixtures/normal.env")
		viper.SetConfigType("env")

		err := viper.ReadInConfig()
		if err != nil {
			b.Fatal(err)
		}

		_ = viper.GetString("S3_BUCKET")
		_ = viper.GetString("SECRET_KEY")
		_ = viper.GetInt("PRIORITY_LEVEL")
	}
}

func BenchmarkViper_Default(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		viper.SetConfigFile("../fixtures/normal.env")
		viper.SetConfigType("env")

		err := viper.ReadInConfig()
		if err != nil {
			b.Fatal(err)
		}

		_ = viper.GetString("S3_BUCKET")
		_ = viper.GetString("SECRET_KEY")
		_ = viper.GetInt("PRIORITY_LEVEL")
	}
}
