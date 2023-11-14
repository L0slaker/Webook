package article

type Article struct {
	Id int64 `gorm:"primaryKey;autoIncrement" bson:"id,omitempty"`
	// 标题长度限制在 1024
	Title string `gorm:"type=varchar(1024)"  bson:"title,omitempty"`
	// BLOB 用于存储大文本数据
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 1.如何设计索引：对于作者来说，应该是查询草稿箱，里面就会有很多篇文章,
	// 还可以考虑按照 "创建时间"/"更新时间" 正序或倒序进行排序
	// 2.查询草稿箱： SELECT * FROM articles WHERE author_id = ?
	// 3.查询指定文章：SELECT * FROM articles WHERE id = 123 ORDER BY `create_time` DESC;
	// 4.最佳选择：在 author_id 和 create_time 上创建联合索引
	//AuthorId   int64 `gorm:"index=aid_ctime"`
	//CreateTime int64 `gorm:"index=aid_ctime"`
	AuthorId   int64 `gorm:"index" bson:"author_id,omitempty"`
	CreateTime int64 `bson:"create_time,omitempty"`
	UpdateTime int64 `bson:"update_time,omitempty"`
	Status     uint8 `bson:"status,omitempty"`
	//DeleteTime int64
}

// PublishedArticle 线上表
type PublishedArticle Article

type PublishedArticleV1 struct {
	Id         int64  `gorm:"primaryKey;autoIncrement" bson:"id,omitempty"`
	Title      string `gorm:"type=varchar(1024)"  bson:"title,omitempty"`
	AuthorId   int64  `gorm:"index" bson:"author_id,omitempty"`
	CreateTime int64  `bson:"create_time,omitempty"`
	UpdateTime int64  `bson:"update_time,omitempty"`
	Status     uint8  `bson:"status,omitempty"`
}

//func (u *model) BeforeCreate(tx *gorm.DB) error {
//	startTime := time.Now()
//	// 将带有键的存储值设置到当前数据库实例的上下文中
//	tx.Set("start_time", startTime)
//	log := slog.Make(sloghuman.Sink(os.Stdout))
//	log.Info(context.Background(), "这是 BeforeCreate 钩子函数")
//	return nil
//}
//
//func (u *model) AfterCreate(tx *gorm.DB) error {
//	val, _ := tx.Get("start_time")
//	startTime, ok := val.(time.Time)
//	if !ok {
//		return nil
//	}
//	// 执行时间
//	duration := time.Since(startTime)
//	log := slog.Make(sloghuman.Sink(os.Stdout))
//	log.Info(context.Background(), "这是 AfterCreate 钩子函数")
//	return nil
//}
//
//type model struct{}
