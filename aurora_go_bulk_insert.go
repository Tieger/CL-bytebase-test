package main

import (
	"crypto/tls"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
	"strings"
	"io/ioutil"
	"log"
	"github.com/go-sql-driver/mysql"
)

const (
	DBHost     = "10.20.7.91"
	//DBHost     = "account-dev.cluster-c9ntc7skbu5q.rds.cn-northwest-1.amazonaws.com.cn"
	DBUser     = "admin"
	DBPassword = "admin123"
	DBName     = "account_test"
	BatchSize  = 1000
	TotalUsers = 100000000
)

func main() {
    // 读取 Aurora 服务器 CA
    pem, err := ioutil.ReadFile("/data/dba/global-bundle.pem")
    if err != nil {
        log.Fatalf("Failed to load Aurora server CA: %s", err)
    }

    // 创建一个新的 CA pool 并添加服务器 CA
    rootCertPool := x509.NewCertPool()
    if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
        log.Fatal("Failed to append PEM.")
    }

    // 创建一个使用服务器 CA 的 TLS 配置
    mysql.RegisterTLSConfig("aurora", &tls.Config{
        RootCAs: rootCertPool,
    })

	// 链接到数据库
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=aurora", DBUser, DBPassword, DBHost, DBName))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	
	mobileBase := 10000000000
	for i := 0; i < TotalUsers; i += BatchSize {
		var values []string //
		for j := 0; j < BatchSize; j++ {
			mobile := mobileBase + i + j //递增手机号
			mobileBidx := sha256.Sum256([]byte(fmt.Sprint(mobile)))
			password := fmt.Sprintf("%x", sha256.Sum256([]byte(randString(32))))
			personID := rand.Intn(1000000) + 1
			regOrigin := "website"
			regIP := "127.0.0.1"
			createdAt := time.Now().Format("2006-01-02 15:04:05")
			updatedAt := createdAt

			value := fmt.Sprintf("UNHEX('%x'), '%d', '%s', '%d', '%s', INET6_ATON('%s'), '%s', '%s'",
				mobileBidx, mobile, password, personID, regOrigin, regIP, createdAt, updatedAt)
			values = append(values, "("+value+")")
		}

		sql := "INSERT INTO user (mobile_bidx, mobile, password, person_id, reg_origin, reg_ip, created_at, updated_at) VALUES " + strings.Join(values, ",")
		_, err = db.Exec(sql)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
