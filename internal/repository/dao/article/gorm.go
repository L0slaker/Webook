package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.CreateTime = now
	art.UpdateTime = now

	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.UpdateTime = now

	res := dao.db.WithContext(ctx).Model(&art).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":       art.Title,
			"content":     art.Content,
			"status":      art.Status,
			"update_time": art.UpdateTime,
		})
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		// 日志
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d", art.Id, art.AuthorId)
	}

	return res.Error
}

func (dao *GORMArticleDAO) GetByAuthorId(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	var arts []Article
	// SELECT * FROM XXX WHERE XX ORDER BY aaa
	// 在设计 order by 的时候要注意让 order by 中的数据命中索引
	// 在同时命中的情况下，可以在磁盘中排序，而不需要拿出数据后再排序
	// 在这里也就是建立一个 author_id,update_time 的联合索引
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", author).
		Limit(limit).Offset(offset).
		//Order("update_time DESC").
		// 按照 "创建时间" 的升序和 "更新时间" 的降序来排序
		Order(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "create_time"}, Desc: false},
			{Column: clause.Column{Name: "update_time"}, Desc: true},
		}}).
		Find(&arts).Error
	return arts, err
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).
		First(&art).Error
	return art, err
}

func (dao *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pub PublishedArticle
	err := dao.db.WithContext(ctx).Model(&PublishedArticle{}).Where("id = ?", id).
		First(&pub).Error
	return pub, err
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 先操作制作库（此时是表），再操作线上库（此时是表）
	// 在事务内，这里采用了闭包形态，GORM 帮我们管理事务的生命周期
	// 也就是说，Begin、Rollback和 Commit 我们都不需要操心
	tx := dao.db.WithContext(ctx).Begin()
	now := time.Now().UnixMilli()
	defer tx.Rollback()
	txDAO := NewGORMArticleDAO(tx)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = txDAO.Insert(ctx, art)
	} else {
		err = txDAO.UpdateById(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	publishArt := PublishedArticle(art)
	publishArt.UpdateTime = now
	publishArt.CreateTime = now
	err = tx.Clauses(clause.OnConflict{
		// 哪些列冲突
		//Columns: nil,
		// 数据冲突并符合WHERE条件的，就会执行 DO UPDATE
		//Where:        clause.Where{},
		// 数据冲突后啥也不做
		//DoNothing: false,
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":       art.Title,
			"content":     art.Content,
			"status":      art.Status,
			"update_time": now,
		}),
	}).Create(&publishArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, tx.Error
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
			Updates(map[string]any{
				"status":      art.Status,
				"update_time": now,
			})
		if res.Error != nil {
			// 数据库有问题
			return res.Error
		}
		if res.RowsAffected != 1 {
			// 要么是 id 有误，要么是作者有误
			// 第二种情况要注意可能有人入侵系统，没必要再用 Id 搜索数据库来区分这两种情况
			// 用 prometheus 打点，只要频繁出现，就触发告警，如何手工介入排查
			return fmt.Errorf("勿操作他人的文章，user_id：%d，author_id：%d", art.Id, art.AuthorId)
		}

		res = tx.Model(&PublishedArticle{}).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
			Updates(map[string]any{
				"status":      art.Status,
				"update_time": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("勿操作他人的文章，user_id：%d，author_id：%d", art.Id, art.AuthorId)
		}
		return nil
	})
}
