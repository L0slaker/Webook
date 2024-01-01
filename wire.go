//go:build wireinject

package main

import (
	"Prove/webook/interactive/events"
	repository2 "Prove/webook/interactive/repository"
	cache2 "Prove/webook/interactive/repository/cache"
	dao2 "Prove/webook/interactive/repository/dao"
	service2 "Prove/webook/interactive/service"
	"Prove/webook/internal/events/article"
	"Prove/webook/internal/repository"
	artRepo "Prove/webook/internal/repository/article"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	artDAO "Prove/webook/internal/repository/dao/article"
	"Prove/webook/internal/service"
	"Prove/webook/internal/web"
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/ioc"
	"github.com/google/wire"
)

var (
	// 第三方依赖
	thirdProvider = wire.NewSet(
		ioc.InitDB, ioc.InitRedis, ioc.InitRLockClient,
		ioc.InitLogger, ioc.InitKafka,
		ioc.NewConsumers, ioc.NewSyncProducer,
	)

	// 用户模块
	userProvider = wire.NewSet(
		dao.NewUserInfoDAO, cache.NewUserCache,
		repository.NewUserInfoRepository,
		service.NewUserService,
	)

	// 验证码模块
	codeProvider = wire.NewSet(
		cache.NewRedisCodeCache,
		repository.NewCodeRepository,
		service.NewCodeService,
	)

	// 文章模块
	articleProvider = wire.NewSet(
		artDAO.NewGORMArticleDAO,
		cache.NewRedisArticleCache,
		artRepo.NewArticleRepository,
		service.NewArticleService,
	)

	// 阅读计数模块
	interProvider = wire.NewSet(
		dao2.NewGORMInteractiveDAO,
		cache2.NewRedisInteractiveCache,
		repository2.NewCachedInteractiveRepository,
		service2.NewInteractiveService,
	)

	// 排行榜模块
	rankingProvider = wire.NewSet(
		cache.NewRankingRedisCache,
		cache.NewRankingLocalCache,
		repository.NewCachedRankingRepository,
		service.NewBatchRankingService,
	)
	// App (*cron.Cron)
	// InitJobs(*job.RankingJob) *cron.Cron
	// InitRankingJob(svc service.RankingService) *job.RankingJob
	// NewBatchRankingService(
	//       ArticleService,
	//       repository.RankingRepository,
	//		 interv1.InteractiveServiceClient) RankingService
	// NewCachedRankingRepository(
	//		 *cache.RankingRedisCache,
	//		 *cache.RankingLocalCache) RankingRepository
	// NewRankingRedisCache() *RankingRedisCache
	// NewRankingLocalCache() *RankingLocalCache
)

func InitWebServer() *App {
	wire.Build(
		thirdProvider,
		userProvider,
		codeProvider,
		articleProvider,
		interProvider,
		rankingProvider,

		// 批量处理
		events.NewInteractiveReadEventBatchConsumer,
		article.NewKafkaProducer,

		ioc.InitSMSService, ioc.InitWechatService,
		ioc.InitInteractiveGRPCClient,
		ioc.InitRankingJob, ioc.InitJobs,

		// 初始化 handler
		web.NewUserHandler, web.NewArticleHandler,
		web.NewOAuth2WechatHandler, ijwt.NewRedisJWT,
		ioc.InitWechatHandlerConfig,

		ioc.InitMiddlewares, ioc.InitEngine,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
