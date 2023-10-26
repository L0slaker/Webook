package integration

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/integration/startup"
	"Prove/webook/internal/repository/dao/article"
	ijwt "Prove/webook/internal/web/jwt"
	"bytes"
	"context"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type ArticleMongodbTestSuite struct {
	suite.Suite
	server  *gin.Engine
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (a *ArticleMongodbTestSuite) SetupSuite() {
	a.server = gin.Default()
	a.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			UserId: 123,
		})
		ctx.Next()
	})
	a.mdb = startup.InitMongoDB()
	node, err := snowflake.NewNode(1)
	assert.NoError(a.T(), err)
	err = article.InitCollections(a.mdb)
	if err != nil {
		panic(err)
	}
	a.col = a.mdb.Collection("articles")
	a.liveCol = a.mdb.Collection("published_articles")
	handler := startup.InitArticleHandler(article.NewMongoDBDAO(a.mdb, node))
	handler.RegisterRoutes(a.server)
}

func (a *ArticleMongodbTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	_, err := a.mdb.Collection("articles").DeleteMany(ctx, bson.D{})
	_, err = a.mdb.Collection("published_articles").DeleteMany(ctx, bson.D{})
	assert.NoError(a.T(), err)
}

func (a *ArticleMongodbTestSuite) TestArticleHandler_Edit() {
	t := a.T()
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		art      Article
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name:   "保存帖子成功",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				// 验证数据
				var art article.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&art)
				assert.NoError(t, err)
				// 确保已经生成了主键
				assert.True(t, art.Id > 0)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)
				art.Id = 0
				art.CreateTime = 0
				art.UpdateTime = 0
				assert.Equal(t, article.Article{
					Title:    "出师表",
					Content:  "愿陛下拖成以讨贼兴复之效，已报先帝之明",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnPublished.ToUint8(),
				}, art)
			},
			art: Article{
				Title:   "出师表",
				Content: "愿陛下拖成以讨贼兴复之效，已报先帝之明",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "保存成功！",
				Data: 1,
			},
		},
		{
			name: "修改已有帖子，并保存！",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				_, err := a.col.InsertOne(ctx, &article.Article{
					Id:         2,
					Title:      "出师表",
					Content:    "愿陛下拖成以讨贼兴复之效，已报先帝之明",
					AuthorId:   123,
					CreateTime: 123,
					UpdateTime: 234,
					Status:     domain.ArticleStatusUnPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				var art article.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.UpdateTime > 234)
				art.UpdateTime = 0
				assert.Equal(t, article.Article{
					Id:         2,
					AuthorId:   123,
					Title:      "李广",
					Content:    "但使龙城飞将在，不教胡马度阴山",
					CreateTime: 123,
					Status:     domain.ArticleStatusUnPublished.ToUint8(),
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "李广",
				Content: "但使龙城飞将在，不教胡马度阴山",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "保存成功！",
				Data: 2,
			},
		},
		{
			name: "修改别人的帖子！",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				_, err := a.col.InsertOne(ctx, &article.Article{
					Id:         3,
					Title:      "李广",
					Content:    "但使龙城飞将在，不教胡马度阴山",
					AuthorId:   789,
					CreateTime: 123,
					UpdateTime: 234,
					Status:     domain.ArticleStatusUnPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				var art article.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, article.Article{
					Id:         3,
					Title:      "李广",
					Content:    "但使龙城飞将在，不教胡马度阴山",
					AuthorId:   789,
					CreateTime: 123,
					UpdateTime: 234,
					Status:     domain.ArticleStatusUnPublished.ToUint8(),
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "明日歌",
				Content: "明日复明日，明日何其多",
			},
			wantCode: http.StatusInternalServerError,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误！",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			a.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}

			var res Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes.Code, res.Code)
			// 只能判定有ID，雪花算法无法确定具体的值
			if tc.wantRes.Data > 0 {
				assert.True(t, res.Data > 0)
			}
			tc.after(t)
		})
	}
}

func (a *ArticleMongodbTestSuite) TestArticleHandler_Publish() {
	t := a.T()
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		art      Article
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name:   "新建帖子并发表",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				// 验证数据
				var art article.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&art)
				assert.NoError(t, err)
				// 确保已经生成了主键
				assert.True(t, art.Id > 0)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)
				art.Id = 0
				art.CreateTime = 0
				art.UpdateTime = 0
				assert.Equal(t, article.Article{
					Title:    "李广",
					Content:  "但使龙城飞将在，不教胡马度阴山",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
				// 验证线上表的数据
				var publishedArt article.PublishedArticle
				err = a.liveCol.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.True(t, publishedArt.Id > 0)
				assert.True(t, publishedArt.CreateTime > 0)
				assert.True(t, publishedArt.UpdateTime > 0)
				publishedArt.Id = 0
				publishedArt.CreateTime = 0
				publishedArt.UpdateTime = 0
				assert.Equal(t, article.PublishedArticle{
					Title:    "李广",
					Content:  "但使龙城飞将在，不教胡马度阴山",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, publishedArt)
			},
			art: Article{
				Title:   "李广",
				Content: "但使龙城飞将在，不教胡马度阴山",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "发布成功！",
				Data: 1,
			},
		},
		{
			// 制作库有，但线上库没有
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				_, err := a.col.InsertOne(ctx, &article.Article{
					Id:         2,
					Title:      "李广",
					Content:    "但使龙城飞将在，不教胡马度阴山",
					AuthorId:   123,
					CreateTime: 456,
					UpdateTime: 789,
					Status:     domain.ArticleStatusUnPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				// 验证数据
				var art article.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				assert.NoError(t, err)
				// 更新时间变了,创建时间没变
				assert.True(t, art.UpdateTime > 789)
				art.UpdateTime = 0
				assert.Equal(t, article.Article{
					Id:         2,
					Title:      "出师表",
					Content:    "愿陛下拖成以讨贼兴复之效，已报先帝之明",
					AuthorId:   123,
					CreateTime: 456,
					Status:     domain.ArticleStatusPublished.ToUint8(),
				}, art)
				// 验证线上表的数据
				var publishedArt article.PublishedArticle
				err = a.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.True(t, publishedArt.CreateTime > 0)
				assert.True(t, publishedArt.UpdateTime > 0)
				publishedArt.CreateTime = 0
				publishedArt.UpdateTime = 0
				assert.Equal(t, article.PublishedArticle{
					Id:       2,
					Title:    "出师表",
					Content:  "愿陛下拖成以讨贼兴复之效，已报先帝之明",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, publishedArt)
			},
			art: Article{
				Id:      2,
				Title:   "出师表",
				Content: "愿陛下拖成以讨贼兴复之效，已报先帝之明",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "发布成功！",
				Data: 2,
			},
		},
		{
			// 制作库和线上库都有
			name: "更新帖子，并重新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				art := article.Article{
					Id:         3,
					Title:      "李广",
					Content:    "但使龙城飞将在，不教胡马度阴山",
					AuthorId:   123,
					CreateTime: 456,
					UpdateTime: 789,
					Status:     domain.ArticleStatusPublished.ToUint8(),
				}
				_, err := a.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				part := article.PublishedArticle(art)
				_, err = a.liveCol.InsertOne(ctx, &part)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				// 验证数据
				var art article.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				assert.NoError(t, err)
				// 更新时间变了，创建时间没变
				assert.True(t, art.UpdateTime > 789)
				art.UpdateTime = 0
				assert.Equal(t, article.Article{
					Id:         3,
					Title:      "出师表",
					Content:    "愿陛下拖成以讨贼兴复之效，已报先帝之明",
					AuthorId:   123,
					CreateTime: 456,
					Status:     domain.ArticleStatusPublished.ToUint8(),
				}, art)

				var publishedArt article.PublishedArticle
				err = a.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&publishedArt)
				assert.NoError(t, err)
				// 更新时间变了，创建时间也变了
				assert.True(t, publishedArt.UpdateTime > 0)
				assert.True(t, publishedArt.CreateTime > 0)
				publishedArt.UpdateTime = 0
				assert.Equal(t, article.PublishedArticle{
					Id:         3,
					Title:      "出师表",
					Content:    "愿陛下拖成以讨贼兴复之效，已报先帝之明",
					AuthorId:   123,
					CreateTime: 456,
					Status:     domain.ArticleStatusPublished.ToUint8(),
				}, publishedArt)
			},
			art: Article{
				Id:      3,
				Title:   "出师表",
				Content: "愿陛下拖成以讨贼兴复之效，已报先帝之明",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "发布成功！",
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				art := article.Article{
					Id:      4,
					Title:   "李广",
					Content: "但使龙城飞将在，不教胡马度阴山",
					// 这个 AuthorId 是其他人的Id
					AuthorId:   789,
					CreateTime: 456,
					UpdateTime: 789,
					Status:     domain.ArticleStatusPublished.ToUint8(),
				}
				_, err := a.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				part := article.PublishedArticle(art)
				_, err = a.liveCol.InsertOne(ctx, &part)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				// 验证数据
				var art article.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&art)

				assert.NoError(t, err)
				assert.Equal(t, "李广", art.Title)
				assert.Equal(t, "但使龙城飞将在，不教胡马度阴山", art.Content)
				assert.Equal(t, int64(789), art.AuthorId)
				assert.Equal(t, int64(456), art.CreateTime)
				assert.Equal(t, int64(789), art.UpdateTime)

				// 更新失败了，数据没有发生变化
				var publishedArt article.PublishedArticle
				err = a.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, "李广", publishedArt.Title)
				assert.Equal(t, "但使龙城飞将在，不教胡马度阴山", publishedArt.Content)
				assert.Equal(t, int64(789), publishedArt.AuthorId)
				assert.Equal(t, int64(456), publishedArt.CreateTime)
				assert.Equal(t, int64(789), publishedArt.UpdateTime)
			},
			art: Article{
				Id:      4,
				Title:   "出师表",
				Content: "愿陛下拖成以讨贼兴复之效，已报先帝之明",
			},
			wantCode: http.StatusInternalServerError,
			wantRes: Result[int64]{
				Msg:  "系统错误！",
				Code: 5,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			a.server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}

			var res Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes.Code, res.Code)
			// 只能判定有ID，雪花算法无法确定具体的值
			if tc.wantRes.Data > 0 {
				assert.True(t, res.Data > 0)
			}
			tc.after(t)
		})
	}
}

func TestMongoArticle(t *testing.T) {
	suite.Run(t, &ArticleMongodbTestSuite{})
}
