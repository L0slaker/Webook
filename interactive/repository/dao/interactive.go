package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error)
	DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}
}

func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	// upsert：因为不确定传入的业务是否为新业务，所以用 upsert 语法
	// 我们可以利用数据库来解决并发问题 -> update a = a+1
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"read_cnt":    gorm.Expr("read_cnt + 1"),
			"update_time": time.Now().UnixMilli(),
		}),
	}).Create(&Interactive{
		BizId:      bizId,
		Biz:        biz,
		ReadCnt:    1,
		CreateTime: now,
		UpdateTime: now,
	}).Error
	// 这种处理方式会有并发问题，如果有两个线程同时进来
	// inter.ReadCnt = 10，那么结果本来应该为12，
	// 但是这种处理可能会使结果为11
	//var inter Interactive
	//err := dao.db.WithContext(ctx).Where("biz_id = ? AND biz = ?", bizId, biz).
	//	First(&inter).Error
	//if err != nil {
	//	return err
	//}
	//cnt := inter.ReadCnt + 1
	//dao.db.WithContext(ctx).Where("biz_id = ? AND biz = ?", bizId, biz).
	//	Updates(map[string]any{
	//		"read_cnt": cnt,
	//	})
}

func (dao *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	// 批量处理消息，要么都成功，要么都失败
	// 为什么批量操作快：
	// 我们只开启了一个事务，等到日志被刷新到磁盘上时，
	// 可以一次性刷新到磁盘上，要比单次处理刷 N 次到磁盘上快
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDAO := NewGORMInteractiveDAO(tx)
		for i := range bizs {
			err := txDAO.IncrReadCnt(ctx, bizs[i], ids[i])
			if err != nil {
				// 阅读计数多一个或少一个无关痛痒，或者 return err也OK
				return err
			}
		}
		return nil
	})
}

func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先插入点赞记录，要考虑是否已经是点赞过的记录
		// 考虑使用 upsert 的语法来处理
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"update_time": now,
				"status":      1,
			}),
		}).Create(&UserLikeBiz{
			Uid:        uid,
			BizId:      bizId,
			Biz:        biz,
			CreateTime: now,
			UpdateTime: now,
			Status:     1,
		}).Error
		if err != nil {
			return err
		}

		// 然后增加点赞计数
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt":    gorm.Expr("like_cnt + 1"),
				"update_time": now,
			}),
		}).Create(&Interactive{
			BizId:      bizId,
			Biz:        biz,
			LikeCnt:    1,
			CreateTime: now,
			UpdateTime: now,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1.软删除点赞记录
		// mysql 会自动调整参数位置，低版本可能不会
		err := tx.Model(&UserLikeBiz{}).Where("uid = ? AND biz_id = ? AND biz = ?",
			uid, bizId, biz).Updates(map[string]any{
			"update_time": now,
			"status":      0,
		}).Error
		if err != nil {
			return err
		}
		// 2.减少点赞数量
		return tx.Model(&Interactive{}).Where("biz_id = ? AND biz = ?",
			bizId, biz).Updates(map[string]any{
			"like_cnt":    gorm.Expr("like_cnt - 1"),
			"update_time": now,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.CreateTime = now
	cb.UpdateTime = now
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 插入收藏项目
		err := dao.db.WithContext(ctx).Create(&cb).Error
		if err != nil {
			return err
		}
		// 更新数量
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt` + 1"),
				"update_time": now,
			}),
		}).Create(&Interactive{
			BizId:      cb.BizId,
			Biz:        cb.Biz,
			CollectCnt: 1,
			CreateTime: now,
			UpdateTime: now,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND uid = ?",
		biz, bizId, uid).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND uid = ?",
		biz, bizId, uid).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ?",
		biz, bizId).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	var res []Interactive
	err := dao.db.WithContext(ctx).Where("biz = ? AND id IN ?",
		biz, ids).Find(&res).Error
	return res, err
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_id_type"`                   // 业务标识符；联合唯一索引
	Biz        string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"` // 联合唯一索引
	ReadCnt    int64  // 阅读计数
	LikeCnt    int64  // 点赞计数
	CollectCnt int64  // 收藏计数
	CreateTime int64
	UpdateTime int64
}

type UserLikeBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 查询用户是否有点赞：WHERE uid = ? AND biz_id = ? AND biz = ?
	// 所以考虑在该三列中创建联合索引，至于 uid 和 biz_id 谁在前，要分场景
	// 1.如果用户要查询自己点赞过哪些，那么 uid 置前
	// WHERE uid = ?
	// 2.如果点赞数量要通过此处来比较/纠正，就 biz_id 和 biz 置前
	// select count(*) where biz = ? and biz_id = ?
	Uid        int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	BizId      int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Biz        string `gorm:"uniqueIndex:uid_biz_id_type;type:varchar(128)"`
	CreateTime int64
	UpdateTime int64
	// 0-代表删除，1-代表有效
	Status uint8
}

type UserCollectionBiz struct {
	Id  int64 `gorm:"primaryKey,autoIncrement"`
	Uid int64 `gorm:"uniqueIndex:biz_type_id_uid"`
	// 收藏夹Id，作为关联关系中的外键，需要索引
	Cid        int64  `gorm:"index"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz        string `gorm:"uniqueIndex:biz_type_id_uid;type:varchar(128)"`
	CreateTime int64
	UpdateTime int64
}

// Collection 收藏夹
type Collection struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type=varchar(1024)"`
	Uid        int64
	CreateTime int64
	UpdateTime int64
}
