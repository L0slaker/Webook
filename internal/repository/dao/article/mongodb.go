package article

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDBDAO struct {
	client  *mongo.Client
	db      *mongo.Database   // webook 的 database
	col     *mongo.Collection // 制作库
	liveCol *mongo.Collection // 线上库
	node    *snowflake.Node   // 雪花算法生成主键
}

func NewMongoDBDAO(db *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBDAO{
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}

func (m *MongoDBDAO) Insert(ctx context.Context, art Article) (int64, error) {
	id := m.node.Generate().Int64()
	now := time.Now().UnixMilli()
	art.Id = id
	art.CreateTime = now
	art.UpdateTime = now
	_, err := m.col.InsertOne(ctx, art)
	return id, err
}

func (m *MongoDBDAO) UpdateById(ctx context.Context, art Article) error {
	filter := bson.M{
		"id":        art.Id,
		"author_id": art.AuthorId,
	}
	update := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title":       art.Title,
		"content":     art.Content,
		"update_time": time.Now().UnixMilli(),
		"status":      art.Status,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		// 日志
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d", art.Id, art.AuthorId)
	}
	return nil
}

func (m *MongoDBDAO) GetByAuthorId(ctx context.Context, authorId int64, offset, limit int) ([]Article, error) {
	//filter := bson.M{
	//	"author_id": authorId,
	//}
	//cursor, err := m.col.Find(ctx,filter)
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 1.保存制作库
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id

	// 2.操作库上线->upsert
	now := time.Now().UnixMilli()
	art.UpdateTime = now
	filter := bson.M{"id": art.Id}
	update := bson.M{
		// 更新，若不存在，则插入
		"$set": PublishedArticle(art),
		// 在插入时插入create_time
		"$setOnInsert": bson.M{"create_time": now},
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return id, err
}

func (m *MongoDBDAO) SyncStatus(ctx context.Context, art Article) error {
	filter := bson.M{
		"id":        art.Id,
		"author_id": art.AuthorId,
	}
	update := bson.M{
		"$set": bson.M{"status": art.Status},
	}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return ErrPossibleIncorrectAuthor
	}
	return nil
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys: bson.D{
				bson.E{Key: "id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "create_time", Value: 1},
			},
			Options: nil,
		},
	}
	_, err := db.Collection("articles").Indexes().CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().CreateMany(ctx, index)
	return err
}
