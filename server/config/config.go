package config

import "os"

type Config struct {
	DBPath     string
	WebhookURL string // 企业微信 webhook 基础地址
	RemindTime string // cron expr, default "0 8 * * *" (每天8点)
	ServerAddr string
}

func Load() *Config {
	return &Config{
		DBPath:     getEnv("DB_PATH", "./family.db"),
		WebhookURL: getEnv("WEBHOOK_URL", "https://qyapi.weixin.qq.com/cgi-bin/webhook/send"),
		RemindTime: getEnv("REMIND_TIME", "0 8 * * *"),
		ServerAddr: getEnv("SERVER_ADDR", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
