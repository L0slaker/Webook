.PHONY: mock
mock:
# mockgen 要指定三个参数：
# 1.source：接口所在的文件
# 2.destination：生成代码的目标路径
# 3.package：生成代码的文件的package

# 在指定路径下执行该Makefile：D:\go\GO-LEARNING\src\Prove
	@mockgen -source=webook\internal\service\user.go -destination=webook\internal\service\mocks\user.mock.go -package=svcmocks
	@mockgen -source=webook\internal\service\code.go -destination=webook\internal\service\mocks\code.mock.go -package=svcmocks
	@mockgen -source=webook\internal\service\article.go -destination=webook\internal\service\mocks\article.mock.go -package=svcmocks
	@mockgen -source=webook\internal\service\interactive.go -destination=webook\internal\service\mocks\interactive.mock.go -package=svcmocks
	@mockgen -source=webook\internal\repository\user.go -destination=webook\internal\repository\mocks\user.mock.go -package=repomocks
	@mockgen -source=webook\internal\repository\code.go -destination=webook\internal\repository\mocks\code.mock.go -package=repomocks
	@mockgen -source=webook\internal\repository\dao\user.go -destination=webook\internal\repository\dao\mocks\user.mock.go -package=daomocks
	@mockgen -source=webook\internal\repository\cache\user.go -destination=webook\internal\repository\cache\mocks\user.mock.go -package=cachemocks
	@mockgen -source=webook\internal\repository\article\article.go -destination=webook\internal\repository\article\mocks\article.mock.go -package=repomocks
	@mockgen -source=webook\internal\repository\article\article_author.go -destination=webook\internal\repository\article\mocks\article_author.mock.go -package=artrepomocks
	@mockgen -source=webook\internal\repository\article\article_reader.go -destination=webook\internal\repository\article\mocks\article_reader.mock.go -package=artrepomocks
	@mockgen -source=webook\pkg\ratelimit\types.go -destination=webook\pkg\ratelimit\mocks\ratelimit.mock.go -package=limitmocks
	@mockgen -source=webook\internal\service\sms\types.go -destination=webook\internal\service\sms\mocks\svc.mock.go -package=smsmocks
	@mockgen -destination=webook\internal\repository\cache\redismocks\cmdable.mock.go -package=redismocks github.com/redis/go-redis/v9 Cmdable
	@go mod tidy

.PHONY: grpc
grpc:
	@buf generate webook/api/proto