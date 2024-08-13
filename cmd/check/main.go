package main

import (
	"flag"
	"flink2mo/conf"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/url"
	"os"
)

var mysqlCnf conf.MysqlConf
var moCnf conf.MysqlConf

func main() {
	cfg := conf.NewConf("./conf/matrixone.ini")
	err := cfg.Load()
	if err != nil {
		panic(err)
	}

	mysqlCnf = conf.MyCnf
	moCnf = conf.MoConf
	var mysqlPwd, moPwd string
	flag.StringVar(&mysqlPwd, "mysqlPwd", "1Qaz2wsx.", "password of mysql database")
	flag.StringVar(&moPwd, "moPwd", "1Qaz2wsx", "password of matrixone database")
	flag.Parse()

	if mysqlPwd == "" || moPwd == "" {
		fmt.Println(fmt.Sprintf("mysql passwprd:%s,mo password:%s", mysqlPwd, moPwd))
		os.Exit(1)
	}

	mysqlConn, err := getDBConn("mysql", mysqlPwd)
	if err != nil {
		os.Exit(1)
	}
	moConn, err := getDBConn("matrixone", moPwd)
	if err != nil {
		os.Exit(1)
	}

	//检查新增、删除是否同步
	table := "mysql_dx"
	var mysqlNum, moNum int64
	mysqlConn.Table(table).Count(&mysqlNum)
	moConn.Table(table).Count(&moNum)
	if mysqlNum != moNum {
		fmt.Println(fmt.Sprintf("[Inconsistent data error]:%s table mysql num %v,mo num %v", table, mysqlNum, moNum))
		os.Exit(1)
	}
	//检查修改是否同步
	type Record struct {
		Id   int64
		Name string
	}
	var list []Record
	err = mysqlConn.Table("modify_record").Find(&list).Error
	if err != nil {
		fmt.Println("查询修改记录数据失败:", err)
		os.Exit(1)
	}
	for _, v := range list {
		var num int64
		err = moConn.Table(table).Where("id=? and name=?", v.Id, v.Name).Count(&num).Error
		if err != nil {
			fmt.Println(fmt.Sprintf("id:%v,name:%s,err:%v", v.Id, v.Name, err))
			os.Exit(1)
		}
		if num == 0 {
			fmt.Println(fmt.Sprintf("id:%v,name:%s", v.Id, v.Name))
			os.Exit(1)
		}
	}

	os.Exit(0)
}

func getDBConn(dataSource string, password string) (*gorm.DB, error) {
	var cnf conf.MysqlConf
	switch dataSource {
	case "mysql":
		cnf = conf.MyCnf
	case "matrixone":
		cnf = conf.MoConf
	}

	username := url.QueryEscape(cnf.Username)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, cnf.HOST, cnf.Port, cnf.DataBase) //MO
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(fmt.Sprintf("%s Database Connection Failed", dataSource)) //Connection failed
		return nil, err
	}
	fmt.Println("Database Connection Succeed") //Connection succeed

	return db, nil
}
