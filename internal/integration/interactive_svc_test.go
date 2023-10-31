package integration

import (
	"Prove/webook/internal/integration/startup"
	"Prove/webook/internal/repository/dao"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
)

type InteractiveTestSuite struct {
	suite.Suite
	db  *gorm.DB
	rdb redis.Cmdable
}

func (s *InteractiveTestSuite) SetupSuite() {
	s.db = startup.InitTestDB()
	s.rdb = startup.InitRedis()
}

func (s *InteractiveTestSuite) TearDownTest() {
	err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_like_bizs`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_collection_bizs`").Error
	assert.NoError(s.T(), err)
}

func (s *InteractiveTestSuite) TestIncrReadCnt() {
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		biz     string
		bizId   int64
		wantErr error
	}{
		{
			name: "增加成功，缓存和数据库中都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         1,
					BizId:      2,
					Biz:        "test",
					ReadCnt:    3,
					LikeCnt:    4,
					CollectCnt: 5,
					CreateTime: 123,
					UpdateTime: 456,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:2", "read_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdateTime > 456)
				data.UpdateTime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					BizId:      2,
					Biz:        "test",
					ReadCnt:    4,
					LikeCnt:    4,
					CollectCnt: 5,
					CreateTime: 123,
				}, data)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "read_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			biz:     "test",
			bizId:   2,
			wantErr: nil,
		},
		{
			// DB 有数据，缓存没有数据
			name: "增加成功，缓存没有数据，数据库中有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         2,
					Biz:        "test",
					BizId:      3,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					CreateTime: 6,
					UpdateTime: 7,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 2).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdateTime > 7)
				data.UpdateTime = 0
				assert.Equal(t, dao.Interactive{
					Id:         2,
					Biz:        "test",
					BizId:      3,
					ReadCnt:    4,
					CollectCnt: 4,
					LikeCnt:    5,
					CreateTime: 6,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:2").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:   "test",
			bizId: 3,
		},
		{
			name:   "增加成功，但缓存和数据库中都没有数据",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("biz_id = ? AND biz = ?", 4, "test").First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.CreateTime > 0)
				assert.True(t, data.UpdateTime > 0)
				assert.True(t, data.Id > 0)
				data.CreateTime = 0
				data.UpdateTime = 0
				data.Id = 0
				assert.Equal(t, dao.Interactive{
					Biz:     "test",
					BizId:   4,
					ReadCnt: 1,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:2").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:   "test",
			bizId: 4,
		},
	}
	// 不同于 AsyncSms，不需要mock，所以创建一个svc就足够
	svc := startup.InitInteractiveService()
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.IncrReadCnt(context.Background(), tc.biz, tc.bizId)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestLike() {
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		uid     int64
		biz     string
		bizId   int64
		wantErr error
	}{
		{
			name: "点赞，数据库和缓存中都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         1,
					BizId:      2,
					Biz:        "test",
					ReadCnt:    3,
					LikeCnt:    4,
					CollectCnt: 5,
					CreateTime: 6,
					UpdateTime: 7,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:2", "like_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdateTime > 7)
				data.UpdateTime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					BizId:      2,
					Biz:        "test",
					ReadCnt:    3,
					LikeCnt:    5,
					CollectCnt: 5,
					CreateTime: 6,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?",
					"test", 2, 123).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.Id > 0)
				assert.True(t, likeBiz.CreateTime > 0)
				assert.True(t, likeBiz.UpdateTime > 0)
				likeBiz.Id = 0
				likeBiz.CreateTime = 0
				likeBiz.UpdateTime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Uid:    123,
					BizId:  2,
					Biz:    "test",
					Status: 1,
				}, likeBiz)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			uid:   123,
			biz:   "test",
			bizId: 2,
		},
	}
	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.Like(context.Background(), tc.biz, tc.bizId, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestDislike() {}

func TestInteractiveService(t *testing.T) {
	suite.Run(t, &InteractiveTestSuite{})
}
