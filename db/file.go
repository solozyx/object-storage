package db

import (
	"database/sql"
	"fmt"

	"github.com/solozyx/object-storage/db/db_mysql"
)

// TableFile : 文件表 tbl_file 结构体
type TableFile struct {
	FileHash string
	// 表示允许为 null
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// 文件上传完成 保存meta
func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	// TODO : NOTICE 预编译语句,防止SQL注入攻击,外部用户写恶意SQL语句,经过逻辑处理嵌入到,DELETE DROP
	stmt, err := db_mysql.DBConn().Prepare(`insert ignore into tbl_file (file_sha1,file_name,file_size,file_addr,status) values (?,?,?,?,1)`)
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	// 判断是否插入了新记录 同1个file_sha1之前可能已经写入了 重复插入会忽略
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			// 虽然SQL执行成功,但是没有产生新的表的记录
			fmt.Printf("File with hash:%s has been uploaded before", filehash)
		}
		return true
	}
	return false
}

// 从mysql获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	// status=1 表示文件存在没被删除
	stmt, err := db_mysql.DBConn().Prepare(`select file_sha1,file_addr,file_name,file_size from tbl_file 
				where file_sha1=? and status=1 limit 1`)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	// 返回当前记录 Scan赋值 字段顺序和SQL语句字段顺序必须相同
	err = stmt.QueryRow(filehash).Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err != nil {
		if err == sql.ErrNoRows {
			// 查不到对应记录 返回参数及错误均为nil
			return nil, nil
		} else {
			fmt.Println(err.Error())
			return nil, err
		}
	}
	return &tfile, nil
}

// 从mysql批量获取文件元信息
func GetFileMetaList(limit int) ([]TableFile, error) {
	stmt, err := db_mysql.DBConn().Prepare(`select file_sha1,file_addr,file_name,file_size from tbl_file where status=1 limit ?`)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cloumns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(cloumns))
	var tfiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}
	fmt.Println(len(tfiles))
	return tfiles, nil
}

//  更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := db_mysql.DBConn().Prepare(`update tbl_file set file_addr=? where file_sha1=? limit 1`)
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("更新文件location失败, filehash:%s", filehash)
		}
		return true
	}
	return false
}
