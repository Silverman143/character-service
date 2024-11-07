package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct{
	Env 				string 			`yaml:"env" env-default:"local"`
	PgSql 				PgSql 			`env-required:"true"`
	Redis 				RedisConfig		`yaml:"cache" env-required:"true"`
	GRPC 				GRPCConfig 		`yaml:"grpc" env-required:"true"`
	Kafka				KafkaConfig		`yaml:"kafka" env-required:"true"`
	Clients				ClientsConfig	`yaml:"clients" `
}

type PgSql struct {
	Host     string `env:"POSTGRES_HOST,required"`
	User     string `env:"POSTGRES_USER,required"`
	Password string `env:"POSTGRES_PASSWORD,required"`
	DbName   string `env:"POSTGRES_DB,required"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	SSLMode  string `env:"POSTGRES_SSLMODE" env-default:"disable"`
}

type GRPCConfig struct {
	Port	int				`yaml:"port"`
	Timeout	time.Duration	`yaml:"timeout"`
}

type KafkaConfig struct{
	TopicRead 		string 		`yaml:"topics_read" env-required:"true"`
	TopicWrite 		string 		`yaml:"topics_write" env-required:"true"`
	GroupID 		string 		`yaml:"group_id" env-required:"true"`
	Broker			[]string	`yaml:"brokers" env-required:"true"`
	User 			string 		`env:"KAFKA_USER,required"`
	Pass 			string 		`env:"KAFKA_PASS,required"`
}

type RedisConfig struct{
	Addr     string `env:"REDIS_ADDR,required"`
	Password string `env:"REDIS_PASSWORD,required"`
	DB       int    `env:"REDIS_DB,required"`
	Lifetime time.Duration `yaml:"lifetime" env-required:"true"`
}

type Client struct {
	Addr			string 			`yaml:"address" env-required:"true"`
	Timeout			time.Duration	`yaml:"timeout" env-required:"true"`
	RetriesCount	int				`yaml:"retries_count" env-required:"true"`
	Insecure 		bool			`yaml:"insecure" env-required:"true"`
}

type ClientsConfig struct {
	User 		Client	`yaml:"user" env-required:"true"`
	Referral 	Client	`yaml:"referral" env-required:"true"`
}

func MustLoad()	*Config {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: No .env file found")
	}

	path := fetchConfigFlag()

	if path == ""{
		panic("config path is empty")
	}

	if _,err := os.Stat(path); os.IsNotExist(err){
		panic("config path dose not exist: " + path)
	}

	var config Config

	if err:=cleanenv.ReadConfig(path, &config); err != nil {
		panic("failed to read config: " + err.Error())
	}

	if err:=cleanenv.ReadEnv(&config); err != nil {
		panic("failed to read env: " + err.Error())
	}

	return &config
}

func fetchConfigFlag() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file") 
	flag.Parse()

	if res == ""{
		res = os.Getenv("CONFIG_PATH")	
	}

	return res
}