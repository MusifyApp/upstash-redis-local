package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"time"
	"upstash-redis-local/internal"
)

var Version = "development"

type Cmd struct {
	RedisAddr string
	Addr      string
	ApiToken  string
	RedisUser string
	RedisPass string
}

func (c *Cmd) Validate() error {
	if c.ApiToken == "" {
		return errors.New("API Token empty")
	}
	if c.RedisAddr == "" {
		return errors.New("redis Addr empty")
	}
	if c.Addr == "" {
		return errors.New("webserver addr empty")
	}
	return nil
}

func main() {
	setupFlags(flag.CommandLine)
	redisAddr := flag.String("redis", ":6379", "Local Redis String")
	redisUser := flag.String("redisusername", "default", "Username for your redis server")
	redisPass := flag.String("redispassword", "", "Password for your redis server")
	addr := flag.String("addr", ":8000", "Local webserver string")
	apiToken := flag.String("token", "upstash", "API token set by user")
	help := flag.Bool("help", false, "")
	flag.Parse()
	cmd := Cmd{
		RedisAddr: *redisAddr,
		RedisUser: *redisUser,
		RedisPass: *redisPass,
		ApiToken:  *apiToken,
		Addr:      *addr,
	}
	if *help {
		printHelp()
		return
	}
	if cmd.Validate() != nil {

	}
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	defer logger.Sync()
	server := internal.Server{Address: cmd.Addr, APIToken: cmd.ApiToken, RedisPool: newPool(cmd.RedisAddr, cmd.RedisUser, cmd.RedisPass), Logger: logger}
	defer server.Serve()
}

func setupFlags(f *flag.FlagSet) {
	f.Usage = func() {
		printHelp()
	}
}

func printHelp() {
	fmt.Printf(`
upstash-redis-local %s
A local server that mimics upstash-redis for local testing purposes!

       * Connect to any local redis of your choice for testing
       * Comlpetely mimics the upstash REST API https://docs.upstash.com/redis/features/restapi

Developed by Hemanth Krishna (https://github.com/DarthBenro008)

USAGE:
	upstash-redis-local
	upstash-redis-local --token upstash --addr :8000 --redis :6379

ARGUMENTS:
	--token			TOKEN		The API token to accept as authorised (default: upstash)
	--addr			ADDR		Address for the server to listen on (default: :8000)
	--redis			ADDR		Address to your redids server (default: :6379)
	--redispassword PASSWORD	Password for your redis server (default: "")
	--redisusername USERNAME	Username for your redis server (default: "")
	--help		Prints this message
`, Version)
}

func connectToRedis(addr string, username string, password string) redis.Conn {
	conn, err := redis.Dial("tcp", addr,
	redis.DialUsername(username),
	redis.DialPassword(password))
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	return conn
}

func newPool(addr string, username string, password string) *redis.Pool {
    return &redis.Pool{
        MaxIdle:     3,
        IdleTimeout: 240 * time.Second,
        Dial: func() (redis.Conn, error) {
            c, err := redis.Dial("tcp", addr,
                redis.DialUsername(username),
                redis.DialPassword(password))
            if err != nil {
                return nil, err
            }
            return c, err
        },
        TestOnBorrow: func(c redis.Conn, t time.Time) error {
            _, err := c.Do("PING")
            return err
        },
    }
}
