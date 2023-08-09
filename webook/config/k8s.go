package config

var KConfig = WebookConfig{
	DB: DBConfig{
		DSN: "root:root@tcp(webook-mysql:3308)/webook",
	},
	Redis: RedisConfig{
		Addr:     "webook-redis:6379",
		DB:       1,
		Password: "",
	},
}
