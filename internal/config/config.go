package config

type Config struct {
	PostgresDSN string `env:"PG_DSN,required"`
	JWTSecret   string `env:"JWT_SECRET,required"`

	MinIOEndpoint      string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
	MinIOAccessKey     string `env:"MINIO_ACCESS_KEY" envDefault:"minioadmin"`
	MinIOSecretKey     string `env:"MINIO_SECRET_KEY" envDefault:"minioadmin"`
	MinIOBucket        string `env:"MINIO_BUCKET" envDefault:"avatars"`
	MinIOUseSSL        bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	MinIOPublicBaseURL string `env:"MINIO_PUBLIC_BASE_URL" envDefault:"http://localhost:9000/avatars"`
}
