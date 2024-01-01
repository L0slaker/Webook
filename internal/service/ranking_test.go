//go:build TODO

package service

import (
	interv1 "Prove/webook/api/proto/gen/inter/v1"
	domain2 "Prove/webook/interactive/domain"
	"Prove/webook/interactive/service"
	"Prove/webook/internal/domain"
	svcmocks "Prove/webook/internal/service/mocks"
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestTopN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (ArticleService, *interv1.InteractiveServiceClient)
		wantArts []domain.Article
		wantErr  error
	}{
		{
			name: "计算成功",
			mock: func(ctrl *gomock.Controller) (ArticleService, service.InteractiveService) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				interSvc := svcmocks.NewMockInteractiveService(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 5).
					Return([]domain.Article{
						{Id: 1, UpdateTime: now, CreateTime: now},
						{Id: 2, UpdateTime: now, CreateTime: now},
						{Id: 3, UpdateTime: now, CreateTime: now},
						{Id: 4, UpdateTime: now, CreateTime: now},
						{Id: 5, UpdateTime: now, CreateTime: now},
					}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 5, 5).
					Return([]domain.Article{}, nil)
				interSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2, 3, 4, 5}).
					Return(map[int64]domain2.Interactive{
						1: {BizId: 1, LikeCnt: 100},
						2: {BizId: 2, LikeCnt: 200},
						3: {BizId: 3, LikeCnt: 300},
						4: {BizId: 4, LikeCnt: 400},
						5: {BizId: 5, LikeCnt: 500},
					}, nil)
				interSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{}).
					Return(map[int64]domain2.Interactive{}, nil)
				return artSvc, interSvc
			},
			wantArts: []domain.Article{
				{Id: 5, UpdateTime: now, CreateTime: now},
				{Id: 4, UpdateTime: now, CreateTime: now},
				{Id: 3, UpdateTime: now, CreateTime: now},
				{Id: 2, UpdateTime: now, CreateTime: now},
				{Id: 1, UpdateTime: now, CreateTime: now},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			artSvc, interSvc := tc.mock(ctrl)
			svc := NewBatchRankingService(artSvc, interSvc).(*BatchRankingService)
			// 为了测试
			svc.batchSize = 5
			svc.n = 5
			svc.scoreFunc = func(t time.Time, likeCnt int64) float64 {
				return float64(likeCnt)
			}

			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
