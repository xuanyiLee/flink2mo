package main

import (
	"database/sql"
	"flag"
	"flink2mo/conf"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Order struct {
	Name      string
	Salary    int
	Age       int
	Entrytime time.Time
	Gender    int
}

var num int
var wg sync.WaitGroup

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func main() {
	cfg := conf.NewConf("./conf/matrixone.ini")
	err := cfg.Load()
	if err != nil {
		fmt.Println("load config error")
		os.Exit(1)
	}

	cnf := conf.MyCnf
	var mysqlPwd string
	flag.StringVar(&mysqlPwd, "mysqlPwd", "1Qaz2wsx.", "password of mysql database")
	flag.IntVar(&num, "num", 10000, "data number of day produce")
	flag.Parse()

	if mysqlPwd == "" {
		fmt.Println(fmt.Sprintf("mysql passwprd:%s", mysqlPwd))
		os.Exit(1)
	}

	username := url.QueryEscape(cnf.Username)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, cnf.Password, cnf.HOST, cnf.Port, cnf.DataBase)
	// 建立数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("create mysql connect err")
		os.Exit(1)
	}
	defer db.Close()

	if num < 100 {
		fmt.Println("data number must be over 100")
		os.Exit(1)
	}

	// 模拟生成数据
	ordersChan := make(chan Order, num)
	generateOrders(ordersChan)
	// 并发插入数据
	for i := 0; i < 10; i++ {
		go insertOrders(db, ordersChan)
	}
	wg.Wait()

	DeleteOrders(db)

	// 删除数据
	ModifyOrders(db)

	os.Exit(0)
}

// 生成模拟数据数据
func generateOrders(ordersChan chan<- Order) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < num; i++ {
		go func() {
			wg.Add(1)
			order := Order{
				Name:      generateRandomString(10, r),
				Salary:    r.Intn(10000) + 2000,
				Age:       r.Intn(13) + 18,
				Entrytime: time.Now().AddDate(-r.Intn(5), -r.Intn(24), -r.Intn(30)),
				Gender:    r.Intn(2),
			}
			ordersChan <- order
		}()
	}

}

// 并发插入数据到数据库
func insertOrders(db *sql.DB, ordersChan <-chan Order) {
	for order := range ordersChan {
		_, err := db.Exec(`INSERT INTO mysql_dx (name,salary,age,entrytime,gender) VALUES (?, ?, ?, ?, ?)`,
			order.Name, order.Salary, order.Age, order.Entrytime, order.Gender)
		if err != nil {
			fmt.Println("插入数据失败:", err)
			os.Exit(1)
		}
		wg.Done()
	}
}

func DeleteOrders(db *sql.DB) {
	total := 0
	total = num / 100
	for i := 0; i < total; i++ {
		_, err := db.Exec("DELETE FROM mysql_dx ORDER BY RAND() LIMIT 1;")
		if err != nil {
			fmt.Println("删除数据失败:", err)
			os.Exit(1)
		}
	}
}

func ModifyOrders(db *sql.DB) {
	db.Query("truncate table modify_record")
	total := 0
	total = num / 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < total; i++ {
		var id int64
		name := generateRandomString(10, r)
		db.QueryRow("select id from mysql_dx ORDER BY RAND() LIMIT 1;").Scan(&id)
		_, err := db.Exec("update mysql_dx set name=? where id=?;", name, id)
		if err != nil {
			fmt.Println("修改数据失败:", err)
			os.Exit(1)
		}
		_, err = db.Exec(`REPLACE INTO modify_record (id,name) VALUES (?, ?)`,
			id, name)
		if err != nil {
			fmt.Println("插入修改记录失败:", err)
			os.Exit(1)
		}
	}
}

func generateRandomString(length int, r *rand.Rand) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
