package service

//func TestArticleService_Save(t *testing.T) {
//	testCases := []struct {
//		name    string
//		mock    func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository)
//		article *domain.Article
//		wantId  int64
//		wantErr error
//	}{
//		{
//			name: "文章已存在，更新文章",
//			mock: func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository) {
//				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
//				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
//				authorRepo.EXPECT().Update(gomock.Any(), &domain.Article{
//					Id:      1,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(nil)
//				return authorRepo, readerRepo
//			},
//			article: &domain.Article{
//				Id:      1,
//				Title:   "我的标题",
//				Content: "我的内容",
//				Author: domain.Author{
//					Id: 123,
//				},
//			},
//			wantId:  1,
//			wantErr: nil,
//		},
//		{
//			name: "文章已存在，但更新失败",
//			mock: func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository) {
//				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
//				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
//				authorRepo.EXPECT().Update(gomock.Any(), &domain.Article{
//					Id:      1,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(errors.New("update error in Save"))
//				return authorRepo, readerRepo
//			},
//			article: &domain.Article{
//				Id:      1,
//				Title:   "我的标题",
//				Content: "我的内容",
//				Author: domain.Author{
//					Id: 123,
//				},
//			},
//			wantId:  1,
//			wantErr: errors.New("update error in Save"),
//		},
//		{
//			name: "文章不存在，创建文章成功！",
//			mock: func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository) {
//				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
//				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
//				authorRepo.EXPECT().Create(gomock.Any(), &domain.Article{
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(int64(2), nil)
//				return authorRepo, readerRepo
//			},
//			article: &domain.Article{
//				Title:   "我的标题",
//				Content: "我的内容",
//				Author: domain.Author{
//					Id: 123,
//				},
//			},
//			wantId:  2,
//			wantErr: nil,
//		},
//		{
//			name: "文章不存在，创建文章失败！",
//			mock: func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository) {
//				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
//				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
//				authorRepo.EXPECT().Create(gomock.Any(), &domain.Article{
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(int64(0), errors.New("create error in Save"))
//				return authorRepo, readerRepo
//			},
//			article: &domain.Article{
//				Title:   "我的标题",
//				Content: "我的内容",
//				Author: domain.Author{
//					Id: 123,
//				},
//			},
//			wantId:  0,
//			wantErr: errors.New("create error in Save"),
//		},
//	}
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//
//			author, reader := tc.mock(ctrl)
//			svc := NewArticleService(author, reader, &logger.NopLogger{})
//			id, err := svc.Save(context.Background(), tc.article)
//			assert.Equal(t, tc.wantId, id)
//			assert.Equal(t, tc.wantErr, err)
//		})
//	}
//}
//
//func TestArticleService_Publish(t *testing.T) {
//	testCases := []struct {
//		name    string
//		mock    func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository)
//		article *domain.Article
//		wantId  int64
//		wantErr error
//	}{
//		{
//			name: "发布成功！",
//			mock: func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository) {
//				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
//				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
//				authorRepo.EXPECT().Create(gomock.Any(), &domain.Article{
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(int64(1), nil)
//				readerRepo.EXPECT().Save(gomock.Any(), &domain.Article{
//					Id:      1,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(int64(1), nil)
//				return authorRepo, readerRepo
//			},
//			article: &domain.Article{
//				Title:   "我的标题",
//				Content: "我的内容",
//				Author: domain.Author{
//					Id: 123,
//				},
//			},
//			wantId:  1,
//			wantErr: nil,
//		},
//		{
//			name: "修改并发表成功！",
//			mock: func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository) {
//				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
//				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
//				authorRepo.EXPECT().Update(gomock.Any(), &domain.Article{
//					Id:      2,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(nil)
//				readerRepo.EXPECT().Save(gomock.Any(), &domain.Article{
//					Id:      2,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(int64(2), nil)
//				return authorRepo, readerRepo
//			},
//			article: &domain.Article{
//				Id:      2,
//				Title:   "我的标题",
//				Content: "我的内容",
//				Author: domain.Author{
//					Id: 123,
//				},
//			},
//			wantId:  2,
//			wantErr: nil,
//		},
//		{
//			name: "保存到制作库失败！",
//			mock: func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository) {
//				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
//				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
//				authorRepo.EXPECT().Update(gomock.Any(), &domain.Article{
//					Id:      3,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(errors.New("save error in authorDB"))
//				return authorRepo, readerRepo
//			},
//			article: &domain.Article{
//				Id:      3,
//				Title:   "我的标题",
//				Content: "我的内容",
//				Author: domain.Author{
//					Id: 123,
//				},
//			},
//			wantId:  0,
//			wantErr: errors.New("save error in authorDB"),
//		},
//		{
//			name: "保存到制作库成功，重试到线上库成功！",
//			mock: func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository) {
//				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
//				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
//				authorRepo.EXPECT().Update(gomock.Any(), &domain.Article{
//					Id:      4,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(nil)
//				readerRepo.EXPECT().Save(gomock.Any(), &domain.Article{
//					Id:      4,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(int64(0), errors.New("save error in readerDB"))
//				readerRepo.EXPECT().Save(gomock.Any(), &domain.Article{
//					Id:      4,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(int64(4), nil)
//				return authorRepo, readerRepo
//			},
//			article: &domain.Article{
//				Id:      4,
//				Title:   "我的标题",
//				Content: "我的内容",
//				Author: domain.Author{
//					Id: 123,
//				},
//			},
//			wantId:  4,
//			wantErr: nil,
//		},
//		{
//			name: "保存到制作库成功，重试到线上库失败！",
//			mock: func(ctrl *gomock.Controller) (author article.ArticleAuthorRepository, reader article.ArticleReaderRepository) {
//				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
//				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
//				authorRepo.EXPECT().Update(gomock.Any(), &domain.Article{
//					Id:      4,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Return(nil)
//				readerRepo.EXPECT().Save(gomock.Any(), &domain.Article{
//					Id:      4,
//					Title:   "我的标题",
//					Content: "我的内容",
//					Author: domain.Author{
//						Id: 123,
//					},
//				}).Times(3).Return(int64(0), errors.New("save error in readerDB"))
//				return authorRepo, readerRepo
//			},
//			article: &domain.Article{
//				Id:      4,
//				Title:   "我的标题",
//				Content: "我的内容",
//				Author: domain.Author{
//					Id: 123,
//				},
//			},
//			wantId:  0,
//			wantErr: errors.New("save error in readerDB"),
//		},
//	}
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//
//			author, reader := tc.mock(ctrl)
//			svc := NewArticleService(author, reader, &logger.NopLogger{})
//			id, err := svc.PublishV1(context.Background(), tc.article)
//			assert.Equal(t, tc.wantId, id)
//			assert.Equal(t, tc.wantErr, err)
//		})
//	}
//}
