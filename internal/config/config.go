package config

type Config struct {
	PostgresDSN string `env:"PG_DSN,required"`
	JWTSecret   string `env:"JWT_SECRET,required"`

	MinIOEndpoint      string `env:"MINIO_ENDPOINT,required"`
	MinIOAccessKey     string `env:"MINIO_ACCESS_KEY,required"`
	MinIOSecretKey     string `env:"MINIO_SECRET_KEY,required"`
	MinIOBucket        string `env:"MINIO_BUCKET,required"`
	MinIOUseSSL        bool   `env:"MINIO_USE_SSL,required"`
	MinIOPublicBaseURL string `env:"MINIO_PUBLIC_BASE_URL,required"`
}
