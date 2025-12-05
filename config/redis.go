package config

import (
	"flag"
	"fmt"
	"github.com/gomodule/redigo/redis"
)

var connection redis.Conn

func ConnectRedis() redis.Conn {
	connStr := fmt.Sprintf("%s:%s", REDISHost, REDISPort)
	var addr = flag.String("Server", connStr, "Redis server address")

	conn, err := redis.Dial("tcp", *addr)
	if err != nil {
		fmt.Println("Failed to connect to redis-server: ", *addr)
	}
	connection = conn
	return connection
}

func CloseRedis()  {
	connection.Close()
}