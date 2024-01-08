package connpool

import (
	"Prove/webook/interactive/repository/dao"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestPoolTest(t *testing.T) {
	webook, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	require.NoError(t, err)
	err = webook.AutoMigrate(&dao.Interactive{})
	require.NoError(t, err)

	inter, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook_inter"))
	require.NoError(t, err)
	err = inter.AutoMigrate(&dao.Interactive{})
	require.NoError(t, err)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &DoubleWritePool{
			src:     webook.ConnPool,
			dst:     inter.ConnPool,
			pattern: atomicx.NewValueOf(patternSrcFirst),
			l:       nil,
		},
	}))
	require.NoError(t, err)
	t.Log(db)
	err = db.Create(dao.Interactive{
		BizId: 123,
		Biz:   "test",
	}).Error
	require.NoError(t, err)

	// 事务问题
	err = db.Transaction(func(tx *gorm.DB) error {
		e := tx.Create(&dao.Interactive{
			BizId: 123,
			Biz:   "test_tx",
		}).Error
		return e
	})
	require.NoError(t, err)

	err = db.Model(&dao.Interactive{}).Updates(map[string]any{
		"biz_id": 789,
	}).Error
	require.NoError(t, err)

}
