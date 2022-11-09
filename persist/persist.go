package persist

import (
	"database/sql"
	"fmt"
	"log"

	// _ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"

	"github.com/zanshin/webanalytics/dbconf"
	"github.com/zanshin/webanalytics/metrics"
)

func Db(dbConfig dbconf.DbConfig) *sql.DB {
	fmt.Println("reached func DB in persist")
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			dbConfig.User, dbConfig.Pass, dbConfig.Host, dbConfig.Port, dbConfig.Name))

	fmt.Println("sql.Open completed in persist")
	defer db.Close()

	// fmt.Sprintf("user=%s password=%s host=%s dbname=%s", dbConfig.User,
	// 	dbConfig.Pass, dbConfig.Host, dbConfig.Name))
	if err != nil {
		log.Fatal("Unable to connect to the database: ", err)
	}
	return db
}

func SetPageViews(db *sql.DB, pageViews []metrics.PageView) {
	fmt.Println("reached func SetPageViews in persist")
	if len(pageViews) < 1 {
		return
	}
	tx, _ := db.Begin()
	stmt, err := db.Prepare("INSERT INTO page_view(timestamp, url, ip_address, user_agent, screen_height, screen_width) VALUES (NOW(), $1, $2, $3, $4, $5)")
	if err != nil {
		log.Println("Unable to prepare statment for PageView: ", err)
	}
	for k := range pageViews {
		_, err = tx.Stmt(stmt).Exec(
			pageViews[k].URL,
			pageViews[k].IPAddress,
			pageViews[k].UserAgent,
			pageViews[k].ScreenHeight,
			pageViews[k].ScreenWidth,
		)
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
}

func SetHrefClicks(db *sql.DB, hrefClicks []metrics.HrefClick) {
	fmt.Println("reached func SetHrefClicks in persist")
	if len(hrefClicks) < 1 {
		return
	}
	tx, _ := db.Begin()
	stmt, err := db.Prepare("INSERT INTO href_click(timestamp, url, ip_address, href, href_rectangle) VALUES (NOW(), $1, $2, $3, box(point($4,$5), point($6,$7)))")
	if err != nil {
		log.Println("Unable to prepare statment for HrefClick: ", err)
	}
	for k := range hrefClicks {
		tx.Stmt(stmt).Exec(
			hrefClicks[k].URL,
			hrefClicks[k].IPAddress,
			hrefClicks[k].Href,
			hrefClicks[k].HrefTop,
			hrefClicks[k].HrefRight,
			hrefClicks[k].HrefBottom,
			hrefClicks[k].HrefLeft,
		)
	}
	tx.Commit()
}
