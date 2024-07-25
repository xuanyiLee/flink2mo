package main

import (
	"database/sql"
	"flag"
	"flink2mo/conf"
	"fmt"
	"math/rand"
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
	flag.StringVar(&mysqlPwd, "mysqlPwd", "", "password of mysql database")
	flag.Parse()

	if mysqlPwd == "" {
		fmt.Println(fmt.Sprintf("mysql passwprd:%s", mysqlPwd))
		os.Exit(1)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cnf.Username, cnf.Password, cnf.HOST, cnf.Port, cnf.DataBase)
	// 建立数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("create mysql connect err")
		os.Exit(1)
	}
	defer db.Close()

	flag.IntVar(&num, "num", 10000, "data number of day produce")
	flag.Parse()

	// 模拟生成订单
	ordersChan := make(chan Order, num)

	generateOrders(ordersChan)

	// 并发插入订单
	for i := 0; i < 10; i++ {
		go insertOrders(db, ordersChan)
	}

	wg.Wait()
	os.Exit(0)
}

// 生成模拟订单数据
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
				Gender:    r.Intn(1),
			}
			ordersChan <- order
		}()
	}

}

// 并发插入订单到数据库
func insertOrders(db *sql.DB, ordersChan <-chan Order) {
	for order := range ordersChan {
		_, err := db.Exec(`INSERT INTO mysql_dx (name,salary,age,entrytime,gender) VALUES (?, ?, ?, ?, ?)`,
			order.Name, order.Salary, order.Age, order.Entrytime, order.Gender)
		if err != nil {
			fmt.Println("插入订单失败:", err)
		}
		wg.Done()
	}
}

func generateRandomString(length int, r *rand.Rand) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
