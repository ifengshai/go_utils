package main

import (
	"database/sql"
	"github.com/ifengshai/go_utils"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 连接数据库
	db, err := sql.Open("mysql", "dev_users:3NiOj0XFw29bHvN5@tcp(192.168.20.40:3306)/zeelool")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 准备 SQL 查询语句
	query := "SELECT *  FROM products INNER JOIN products_groups_products ON products." +
		"id = products_groups_products.product_id  WHERE products_groups_products." +
		"products_group_id = 70  ORDER BY products_groups_products.sort ASC  LIMIT 200 OFFSET 0"

	// 记录开始时间
	start := time.Now()
	timeTemp := make([]int, 0)

	for i := 0; i < 100; i++ {
		// 执行 SQL 查询
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		// 记录结束时间
		elapsed := time.Since(start)

		// 将持续时间转换为毫秒数（float64）
		milliseconds := float64(elapsed.Milliseconds())
		// 将毫秒数转换为 int
		timeTemp = append(timeTemp, int(milliseconds))
		time.Sleep(100 * time.Millisecond)
	}
	var c int
	for _, b := range timeTemp {
		c += b
	}
	d := c / 100

	// 打印 SQL 执行时间
	go_utils.Printf(d)
}
