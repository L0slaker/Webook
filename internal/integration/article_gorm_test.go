package integration

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/integration/startup"
	"Prove/webook/internal/repository/dao/article"
	ijwt "Prove/webook/internal/web/jwt"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ArticleGORMTestSuite 测试套件
type ArticleGORMTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

// SetupSuite 在测试开始前，初始化内容
func (a *ArticleGORMTestSuite) SetupSuite() {
	a.server = gin.Default()
	a.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			UserId: 123,
		})
	})
	a.db = startup.InitTestDB()
	articleHdl := startup.InitArticleHandler(article.NewGORMArticleDAO(a.db))
	// 注册路由
	articleHdl.RegisterRoutes(a.server)
}

// TearDownTest 清空所有数据，并且自增主键h恢复到1
func (a *ArticleGORMTestSuite) TearDownTest() {
	a.db.Exec("TRUNCATE TABLE articles")
	a.db.Exec("TRUNCATE TABLE published_articles")
}

func (a *ArticleGORMTestSuite) TestArticleHandler_Edit() {
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
			name: "保存帖子成功！",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				var art article.Article
				err := a.db.Where("id = ?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)
				art.CreateTime = 0
				art.UpdateTime = 0
				assert.Equal(t, article.Article{
					Id:       1,
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
				a.db.Create(article.Article{
					Id:         2,
					Title:      "出师表",
					Content:    "愿陛下拖成以讨贼兴复之效，已报先帝之明",
					AuthorId:   123,
					CreateTime: 123,
					UpdateTime: 234,
					Status:     domain.ArticleStatusUnPublished.ToUint8(),
				})
			},
			after: func(t *testing.T) {
				var art article.Article
				err := a.db.Where("id = ?", 2).First(&art).Error
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
				err := a.db.Create(article.Article{
					Id:      3,
					Title:   "李广",
					Content: "但使龙城飞将在，不教胡马度阴山",
					// 测试模拟的用户 ID 是123，这里是 789
					// 意味着你在修改别人的数据
					AuthorId:   789,
					CreateTime: 123,
					UpdateTime: 234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art article.Article
				err := a.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, article.Article{
					Id:         3,
					Title:      "李广",
					Content:    "但使龙城飞将在，不教胡马度阴山",
					AuthorId:   789,
					CreateTime: 123,
					UpdateTime: 234,
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
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			a.server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}

			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)

			assert.Equal(t, tc.wantRes, webRes)
			tc.after(t)
		})
	}
}

func (a *ArticleGORMTestSuite) TestArticleHandler_Publish() {
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
				// 验证数据
				var art article.Article
				err := a.db.Where("author_id = ?", 123).First(&art).Error
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
				err = a.db.Where("author_id = ?", 123).First(&publishedArt).Error
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
				err := a.db.Create(&article.Article{
					Id:         2,
					Title:      "李广",
					Content:    "但使龙城飞将在，不教胡马度阴山",
					AuthorId:   123,
					CreateTime: 456,
					UpdateTime: 789,
					Status:     domain.ArticleStatusUnPublished.ToUint8(),
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art article.Article
				err := a.db.Where("id = ?", 2).First(&art).Error
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
				err = a.db.Where("id = ?", 2).First(&publishedArt).Error
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
				art := article.Article{
					Id:         3,
					Title:      "李广",
					Content:    "但使龙城飞将在，不教胡马度阴山",
					AuthorId:   123,
					CreateTime: 456,
					UpdateTime: 789,
					Status:     domain.ArticleStatusPublished.ToUint8(),
				}
				err := a.db.Create(&art).Error
				assert.NoError(t, err)
				part := article.PublishedArticle(art)
				err = a.db.Create(&part).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art article.Article
				err := a.db.Where("id = ?", 3).First(&art).Error
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
				err = a.db.Where("id = ?", 3).First(&publishedArt).Error
				assert.NoError(t, err)
				// 更新时间变了，创建时间也变了
				assert.True(t, publishedArt.UpdateTime > 0)
				assert.True(t, publishedArt.CreateTime > 0)
				publishedArt.UpdateTime = 0
				publishedArt.CreateTime = 0
				assert.Equal(t, article.PublishedArticle{
					Id:       3,
					Title:    "出师表",
					Content:  "愿陛下拖成以讨贼兴复之效，已报先帝之明",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
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
				err := a.db.Create(&art).Error
				assert.NoError(t, err)
				part := article.PublishedArticle(art)

				err = a.db.Create(&part).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art article.Article
				// 更新失败了，数据没有发生变化
				err := a.db.Where("id = ?", 4).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, "李广", art.Title)
				assert.Equal(t, "但使龙城飞将在，不教胡马度阴山", art.Content)
				assert.Equal(t, int64(789), art.AuthorId)
				assert.Equal(t, int64(456), art.CreateTime)
				assert.Equal(t, int64(789), art.UpdateTime)

				// 更新失败了，数据没有发生变化
				var publishedArt article.PublishedArticle
				err = a.db.Where("id = ?", 3).First(&publishedArt).Error
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

			assert.Equal(t, tc.wantRes, res)
			tc.after(t)
		})
	}
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleGORMTestSuite{})
}

func (a *ArticleGORMTestSuite) TestABC() {
	a.T().Log("hello , 这是测试套件")
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Result data字段若是any类型反序列化会出现问题
type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
