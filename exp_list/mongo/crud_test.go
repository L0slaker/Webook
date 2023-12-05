package mongo

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongo_Crud(t *testing.T) {
	// 控制初始化超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	monitor := &event.CommandMonitor{
		// 每个命令（查询）执行之前
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
		// 执行成功
		//Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
		//},
		// 执行失败
		//Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
		//},
	}
	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)
	defer func() {
		// 用完就是要记得关掉。正常来说，都是在应用退出的时候关掉。
		_ = client.Disconnect(context.Background())
	}()

	// 初始化集合
	col := client.Database("webook").Collection("articles")

	// 插入数据
	res, err := col.InsertOne(ctx, Article{
		Id:       123,
		Title:    "诸葛孔明",
		Content:  "愿陛下拖成以讨贼兴复之效",
		AuthorId: 123,
	})
	// mongodb 的文档id，也就是 _id 字段
	fmt.Printf("id：%s", res.InsertedID)

	// bson = binary json
	// 查询 id = 123
	filter := bson.D{bson.E{Key: "id", Value: 123}}
	var art Article
	err = col.FindOne(ctx, filter).Decode(&art)
	assert.NoError(t, err)
	fmt.Printf("%v \n", art)

	art = Article{}
	err = col.FindOne(ctx, Article{Id: 123}).Decode(&art)
	if err == mongo.ErrNoDocuments {
		fmt.Println(err)
	}
	assert.NoError(t, err)
	fmt.Printf("%v \n", art)

	// 更新多条数据
	//sets := bson.D{bson.E{Key: "$set", Value: bson.D{......}}}
	// 更新单条数据
	sets := bson.D{bson.E{Key: "$set", Value: bson.E{Key: "title", Value: "new title"}}}
	updateRes, err := col.UpdateOne(ctx, filter, sets)
	assert.NoError(t, err)
	fmt.Println("affected", updateRes.ModifiedCount)

	// OR
	or := bson.A{
		bson.M{"id": 123},
		bson.M{"id": 456},
	}
	orRes, err := col.Find(ctx, bson.D{bson.E{Key: "$or", Value: or}})
	assert.NoError(t, err)
	var articles []Article
	err = orRes.All(ctx, &articles)
	assert.NoError(t, err)

	// AND
	and := bson.A{
		bson.D{bson.E{
			Key:   "id",
			Value: 123,
		}},
		bson.D{bson.E{
			Key:   "title",
			Value: "我的标题",
		}},
	}
	andRes, err := col.Find(ctx, bson.D{bson.E{Key: "$and", Value: and}})
	assert.NoError(t, err)
	articles = []Article{}
	err = andRes.All(ctx, &articles)
	assert.NoError(t, err)

	// IN
	in := bson.D{bson.E{
		Key:   "id",
		Value: bson.M{"$in": []any{123, 456}},
	}}
	inRes, err := col.Find(ctx, in)
	articles = []Article{}
	err = inRes.All(ctx, &articles)
	assert.NoError(t, err)

	// 查询特定字段
	inRes, err = col.Find(ctx, in, options.Find().SetProjection(bson.M{
		"id":    1,
		"title": 1,
	}))
	articles = []Article{}
	err = inRes.All(ctx, &articles)
	assert.NoError(t, err)

	// 删除
	delRes, err := col.DeleteMany(ctx, filter)
	assert.NoError(t, err)
	fmt.Println("deleted", delRes.DeletedCount)
}

func TestSnowFlake(t *testing.T) {
	// Create a new Node with a Node number of 1
	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Generate a snowflake ID.
	id := node.Generate()

	// Print out the ID in a few different ways.
	fmt.Printf("Int64  ID: %d\n", id)
	fmt.Printf("String ID: %s\n", id)
	fmt.Printf("Base2  ID: %s\n", id.Base2())
	fmt.Printf("Base64 ID: %s\n", id.Base64())

	// Print out the ID's timestamp
	fmt.Printf("ID Time  : %d\n", id.Time())

	// Print out the ID's node number
	fmt.Printf("ID Node  : %d\n", id.Node())

	// Print out the ID's sequence number
	fmt.Printf("ID Step  : %d\n", id.Step())

	// Generate and print, all in one.
	fmt.Printf("ID       : %d\n", node.Generate().Int64())
}

type Article struct {
	Id         int64  `bson:"id,omitempty"`
	Title      string `bson:"title,omitempty"`
	Content    string `bson:"content,omitempty"`
	AuthorId   int64  `bson:"author_id,omitempty"`
	CreateTime int64  `bson:"createTime,omitempty"`
	UpdateTime int64  `bson:"updateTime,omitempty"`
	Status     uint8  `bson:"status,omitempty"`
}
