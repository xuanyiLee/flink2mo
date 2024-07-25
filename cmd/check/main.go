package main

import (
	"flag"
	"flink2mo/conf"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	flag.StringVar(&mysqlPwd, "mysqlPwd", "", "password of mysql database")
	flag.StringVar(&moPwd, "moPwd", "", "password of matrixone database")
	flag.Parse()

	if mysqlPwd == "" || moPwd == "" {
		fmt.Println(fmt.Sprintf("mysql passwprd:%s,mo password:%s", mysqlPwd, moPwd))
		os.Exit(1)
	}
	mysqlCnf.Password, moCnf.Password = mysqlPwd, moPwd

	mysqlConn, err := getDBConn("mysql")
	if err != nil {
		os.Exit(1)
	}

	moConn, err := getDBConn("matrixone")
	if err != nil {
		os.Exit(1)
	}

	//arr := strings.Split(cfg.Tables, ",")
	var mysqlNum, moNum int64
	table := "mysql_dx"
	//for _, table := range arr {
	mysqlConn.Table(table).Count(&mysqlNum)
	moConn.Table(table).Count(&moNum)
	if mysqlNum != moNum {
		fmt.Println(fmt.Sprintf("[Inconsistent data error]:%s table mysql num %v,mo num %v", table, mysqlNum, moNum))
		os.Exit(1)
	}
	mysqlNum = 0
	moNum = 0
	//}

	os.Exit(0)
}

func getDBConn(dataSource string) (*gorm.DB, error) {
	var cnf conf.MysqlConf
	switch dataSource {
	case "mysql":
		cnf = conf.MyCnf
	case "matrixone":
		cnf = conf.MoConf
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cnf.Username, cnf.Password, cnf.HOST, cnf.Port, cnf.DataBase) //MO
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(fmt.Sprintf("%s Database Connection Failed", dataSource)) //Connection failed
		return nil, err
	}
	fmt.Println("Database Connection Succeed") //Connection succeed

	return db, nil
}
