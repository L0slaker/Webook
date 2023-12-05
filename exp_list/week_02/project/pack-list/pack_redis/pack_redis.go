package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

func main() {
	// 1.创建 Redis 客户端连接
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 服务器地址和端口
		Password: "",               // Redis 服务器密码
		DB:       0,                // Redis 数据库索引
	})
	defer client.Close()

	// 执行命令....
	ctx := context.Background()

	// 2.设置键值对
	err := client.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	// 获取键对应的值
	value, err := client.Get(ctx, "key").Result()
	if err == redis.Nil {
		fmt.Println("key does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("字符串-key:", value)
	}

	// 3.设置哈希表字段的值
	err = client.HSet(ctx, "hash", "field", "value").Err()
	if err != nil {
		panic(err)
	}

	// 获取哈希表字段的值
	value, err = client.HGet(ctx, "hash", "field").Result()
	if err == redis.Nil {
		fmt.Println("key does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("哈希表-key:", value)
	}

	// 4.在列表左侧插入元素
	err = client.LPush(ctx, "list", "value1", "value2").Err()
	if err != nil {
		panic(err)
	}

	// 获取列表的长度
	length, err := client.LLen(ctx, "list").Result()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("列表-list length:", length)
	}

	// 读取列表元素范围
	values, err := client.LRange(ctx, "list", 0, -1).Result()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("列表-list values:", values)
	}

	// 5.添加到集合元素
	err = client.SAdd(ctx, "set", "member1", "member2").Err()
	if err != nil {
		panic(err)
	}

	// 获取集合的所有成员
	members, err := client.SMembers(ctx, "set").Result()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("集合-set members:", members)
	}

	// 6.添加带有分数的成员到有序集合
	err = client.ZAdd(ctx, "zset",
		&redis.Z{Score: 1, Member: "member1"},
		&redis.Z{Score: 2, Member: "member2"}).Err()
	if err != nil {
		panic(err)
	}

	members, err = client.ZRange(ctx, "zset", 0, -1).Result()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("有序集合-zset members:", members)
	}

	// 7.创建事务
	tx := client.TxPipeline()

	// 将多个命令添加到事务中
	tx.Incr(ctx, "counter")
	tx.Expire(ctx, "counter", time.Hour)

	// 执行事务并获取结果
	_, err = tx.Exec(ctx)
	if err != nil {
		panic(err)
	}

	// 获取计数器的最新值
	count, err := client.Get(ctx, "counter").Int()
	if err != nil && err != redis.Nil {
		panic(err)
	}
	fmt.Println("counter", count)

	// 8.创建订阅者
	pubSub := client.Subscribe(ctx, "channel1")

	// 接收订阅的消息
	msg, err := pubSub.ReceiveMessage(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Received msg: ", msg.Payload)

	// 9.创建自定义连接池

	// 10.执行Lua脚本
	script := `return redis.call("GET",KEYS[1])`

	result, err := client.Eval(ctx, script, []string{"key"}).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Lua script result: ", result)

	// 11.删除键
	client.Del(ctx, "key1", "key2").Err()

	// 检查键是否存在
	exists, err := client.Exists(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key exists: ", exists)

	// 12.设置键的过期时间
	err = client.Expire(ctx, "key", time.Hour).Err()
	if err != nil {
		panic(err)
	}

	// 获取键的剩余生存时间
	ttl, err := client.TTL(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Key TTL:", ttl)

	// 13.检查连接是否建立
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Ping response:", pong)

	// 关闭连接
	err = client.Close()
	if err != nil {
		panic(err)
	}
}
