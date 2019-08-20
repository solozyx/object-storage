package meta

import (
	"sort"
	"sync"

	mydb "filestore-server/db"
)

// FileMeta : 文件元信息结构
type FileMeta struct {
	// 文件Sha1作为文件唯一标识 也可用MD5
	FileSha1 string
	FileName string
	FileSize int64
	// 文件存储路径
	Location string
	// 上传时间
	UploadAt string
}

// 存储所有上传文件的元信息
var fileMetas map[string]FileMeta
// 互斥锁 保证线程安全
var mu sync.Mutex

// 内建init函数会在首次运行程序时执行1次
func init() {
	fileMetas = make(map[string]FileMeta)
	// 初始化互斥锁 包含共享变量 fileMetas
	mu = sync.Mutex{}
}

// UpdateFileMeta : 新增/更新文件元信息
func UpdateFileMeta(fmeta FileMeta) {
	// 对 userList 加互斥锁
	mu.Lock()
	// 解互斥锁,防止发生死锁
	defer mu.Unlock()

	fileMetas[fmeta.FileSha1] = fmeta
}

// GetFileMeta : 通过sha1值获取文件的元信息对象
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// UpdateFileMetaDB : 新增/更新文件元信息到mysql中
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return mydb.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// GetFileMetaDB : 从mysql获取文件元信息
func GetFileMetaDB(fileSha1 string) (*FileMeta, error) {
	tfile, err := mydb.GetFileMeta(fileSha1)
	if tfile == nil || err != nil {
		return nil, err
	}
	fmeta := FileMeta{
		FileSha1: tfile.FileHash,
		// NullString 封装了String的结构体 -> String
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return &fmeta, nil
}

// GetLastFileMetas : 获取批量的文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}

	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count]
}

// GetLastFileMetasDB : 批量从mysql获取文件元信息
func GetLastFileMetasDB(limit int) ([]FileMeta, error) {
	tfiles, err := mydb.GetFileMetaList(limit)
	if err != nil {
		return make([]FileMeta, 0), err
	}

	tfilesm := make([]FileMeta, len(tfiles))
	for i := 0; i < len(tfilesm); i++ {
		tfilesm[i] = FileMeta{
			FileSha1: tfiles[i].FileHash,
			FileName: tfiles[i].FileName.String,
			FileSize: tfiles[i].FileSize.Int64,
			Location: tfiles[i].FileAddr.String,
		}
	}
	return tfilesm, nil
}

// RemoveFileMeta : 删除文件元信息map内存维护数据
func RemoveFileMeta(fileSha1 string) {
	mu.Lock()
	defer mu.Unlock()
	delete(fileMetas, fileSha1)
}
