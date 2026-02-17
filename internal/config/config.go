package config

type Config struct {
	PostgresDSN string `env:"PG_DSN,required"`
	JWTSecret   string `env:"JWT_SECRET,required"`
}
