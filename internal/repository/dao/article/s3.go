package article

import (
	"Prove/webook/internal/domain"
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

var (
	statusPrivate              = domain.ArticleStatusPrivate.ToUint8()
	ErrPossibleIncorrectAuthor = errors.New("作者可能有误！")
)

type S3DAO struct {
	oss *s3.S3
	GORMArticleDAO
	bucket *string
}

func NewOssDAO(oss *s3.S3, db *gorm.DB) ArticleDAO {
	return &S3DAO{
		oss:    oss,
		bucket: ekit.ToPtr[string]("webook-1314583317"),
		GORMArticleDAO: GORMArticleDAO{
			db: db,
		},
	}
}

func (o *S3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := o.db.Transaction(func(tx *gorm.DB) error {
		var err error
		now := time.Now().UnixMilli()
		// 制作库
		txDAO := NewGORMArticleDAO(tx)
		if id > 0 {
			err = txDAO.UpdateById(ctx, art)
		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		// 线上库不保存 Content，要准备上传到 OSS 里面
		publishArt := PublishedArticleV1{
			Id:         art.Id,
			Title:      art.Title,
			AuthorId:   art.AuthorId,
			CreateTime: now,
			UpdateTime: now,
			Status:     art.Status,
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":       art.Title,
				"update_time": now,
				"status":      art.Status,
			}),
		}).Create(&publishArt).Error
	})
	if err != nil {
		return 0, err
	}
	_, err = o.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      o.bucket,
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=uft-8"),
	})
	// oss操作失败——监控、重试、补偿等等机制
	return id, err
}

func (o *S3DAO) SyncStatus(ctx context.Context, art Article) error {
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
			Update("status", art.Status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		res = tx.Model(&PublishedArticle{}).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
			Update("status", art.Status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		return nil
	})
	if err != nil {
		return err
	}
	if art.Status == statusPrivate {
		_, err = o.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: o.bucket,
			Key:    ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		})
	}
	return err
}
