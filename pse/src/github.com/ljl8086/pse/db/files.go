package db

import (
	_ "github.com/go-sql-driver/mysql"
	. "github.com/ljl8086/pse/utils"
	"github.com/ljl8086/pse/vo"
	cm "github.com/ljl8086/pse/common"
	"time"
)

var (
	err    error
	errStr = "db error:"
	isTaskIng = false
)

func init() {
	go startTimer()
}

//开启定时任务
func startTimer() {
	timer1 := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timer1.C:
			task()
		}
	}
}

//任务
func task() {
	if	isTaskIng { 
		return
	}
	
	isTaskIng = true
	go func() {
		cm.Log.Debugln("Check the validity of the FDFS file...")
		fdbs := getTimeoutFileList()
		lenF := len(fdbs)
		var ids []int
		for i:=0;i<lenF;i++{
			err := cm.FdfsClient.DeleteFile(fdbs[i].FileName)
			if(err!=nil){
				ids = append(ids,fdbs[i].Id)
			}
		}
		deleteFiles(ids)
		isTaskIng = false
	}()
}

func getTimeoutFileList() []vo.FilesDB{
	var fdbs []vo.FilesDB
	stmt, err := cm.Db.Prepare(" INSERT INTO lock_files(files_id,locked_ip) SELECT f.id,? FROM files f LEFT JOIN lock_files lf ON f.id = lf.files_id WHERE f.deadline<UNIX_TIMESTAMP(NOW()) AND lf.files_id IS NULL limit 0,3")
	if makeErr(errStr, err) {
		stmt.Close()
		return fdbs
	}
	_, err = stmt.Exec(cm.CfWebPort)
	stmt.Close()
	if makeErr(errStr, err) {
		return fdbs
	}

	stmt, err = cm.Db.Prepare("SELECT f.id,f.file_name FROM files f JOIN lock_files lf ON f.id = lf.files_id AND lf.locked_ip=?")
	if makeErr(errStr, err) {
		stmt.Close()
		return fdbs
	}

	row, err := stmt.Query(cm.CfWebPort)
	if makeErr(errStr, err) {
		stmt.Close()
		return fdbs
	}

	if row.Next() {
		fdb := vo.FilesDB{}
		row.Scan(&fdb.Id, &fdb.FileName)
		fdbs = append(fdbs,fdb)
	}
	stmt.Close()
	return fdbs
}

func SaveFiles(fdb *vo.FilesDB, mFileName string) {
	errStr := "temp file save to db error:"
	if fdb.IsSlave {
		mainFdb, err := getFiles(mFileName)
		if makeErr(errStr, err) {
			return
		}
		fdb.Deadline = mainFdb.Deadline
		fdb.ParentId = mainFdb.Id
		fdb.FileType = mainFdb.FileType
	}

	stmt, err := cm.Db.Prepare("insert into files(file_name,type,deadline,is_slave,slave_suffix,parent_id) value(?,?,?,?,?,?)")
	defer stmt.Close()
	
	if makeErr(errStr, err) {
		return
	}

	res, err := stmt.Exec(fdb.FileName, fdb.FileType, fdb.Deadline, fdb.IsSlave, fdb.SlaveSuffix, fdb.ParentId)
	if makeErr(errStr, err) {
		return
	}

	_, err = res.RowsAffected()
	if makeErr(errStr, err) {
		return
	}
}

func deleteFiles(ids []int){
	if(len(ids)<=0){ 
		return
	}
	
	errStr = "delete db fail:"
	tx,err := cm.Db.Begin()
	idsStr := JoinInt2(",",ids)
	sql := Join("","delete from files where id in(",idsStr,")")
	stmt,err := cm.Db.Prepare(sql)
	defer stmt.Close()
	
 	if makeErr(errStr, err) {
 		tx.Rollback();
		return
	}
 	_,err = stmt.Exec()
 	 if makeErr(errStr, err) {
 	 	tx.Rollback();
		return
	}
 	
 	stmt,err = cm.Db.Prepare(Join("","delete from lock_files where files_id in(",idsStr,")"))
 	if makeErr(errStr, err) {
 		tx.Rollback();
		return
	}
 	_,err = stmt.Exec()
 	 if makeErr(errStr, err) {
 	 	tx.Rollback();
		return
	}
 	tx.Commit()
}

func getFiles(filename string) (*vo.FilesDB, error) {
	var fdb vo.FilesDB
	stmt, err := cm.Db.Prepare("SELECT id,file_name,type,deadline,is_slave,slave_suffix,parent_id FROM files WHERE file_name=?")
	defer stmt.Close()
	
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(filename)
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		rows.Scan(&fdb.Id, &fdb.FileName, &fdb.FileType, &fdb.Deadline, &fdb.IsSlave, &fdb.SlaveSuffix, &fdb.ParentId)
	}
	return &fdb, nil
}

func makeErr(msg string, err error) bool {
	if err != nil {
		cm.Log.Error(msg, err.Error())
		return true
	}
	return false
}
