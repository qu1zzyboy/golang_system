package snapShot

import (
	"bytes"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var (
	searchPrefixStr  = "snapshot:"
	searchPrefixByte = []byte(searchPrefixStr) // 用于查询的前缀
)

// IndexManager 快照索引管理器,线程安全的
type IndexManager struct {
	db *badger.DB
}

// NewIndexManager 创建一个新的快照索引管理器(默认内存模式)
func NewIndexManager() (*IndexManager, error) {
	opts := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &IndexManager{db: db}, nil
}

// Add 添加一个快照(设置 TTL)
func (s *IndexManager) Add(timeStamp int64, value []byte, ttl time.Duration) error {
	return s.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(s.keyFromTimeStamp(timeStamp), value).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
}

// GetRange 查询某一时间范围内的所有快照
func (s *IndexManager) GetRange(startTs, endTs int64) ([][]byte, error) {
	startKey := s.keyFromTimeStamp(startTs)
	endKey := s.keyFromTimeStamp(endTs)
	var result [][]byte
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = searchPrefixByte // 如果有统一前缀,可以加快扫描
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(startKey); it.Valid(); it.Next() {
			item := it.Item()
			if bytes.Compare(item.Key(), endKey) > 0 {
				break
			}
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			result = append(result, val)
		}
		return nil
	})
	return result, err
}

// DeleteByTimestamp 删除一个快照
func (s *IndexManager) DeleteByTimestamp(ts int64) error {
	key := s.keyFromTimeStamp(ts)
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Close 关闭数据库
func (s *IndexManager) Close() error {
	return s.db.Close()
}

// 构造时间戳 key
func (s *IndexManager) keyFromTimeStamp(ts int64) []byte {
	// 时间戳左填充零以确保排序一致
	return []byte(searchPrefixStr + time.UnixMilli(ts).Format("20060102T150405.000"))
}
