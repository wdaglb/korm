package korm

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

type Test struct {
	Id int64
	User string
}

func TestContext(t *testing.T)  {
	err := Connect(Config{
		Driver: "mysql",
		Host:   "192.168.1.150",
		Port:   3306,
		User:   "root",
		Pass:   "jdj123456_",
		Database: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := &Context{}

	i := 0
	insertData := &Test{
		User: "test",
	}

	err = ctx.Transaction(func() error {
		i++
		insertData.User = fmt.Sprintf("user%v", i)
		if err := ctx.Model(&insertData).Create(); err != nil {
			return fmt.Errorf("create fail: %v\n", err)
		}
		t.Logf("create success\n")

		//if err := ctx.Model(Test{Id: insertData.Id}).Delete(); err != nil {
		//	return fmt.Errorf("delete fail: %v\n", err)
		//}
		//t.Logf("delete success\n")

		_ = ctx.Transaction(func() error {
			i++
			insertData.User = fmt.Sprintf("user%v", i)
			if err := ctx.Model(&insertData).Create(); err != nil {
				return fmt.Errorf("create2 fail: %v\n", err)
			}
			t.Logf("create2 success\n")

			i++
			insertData.User = fmt.Sprintf("user%v", i)
			if err := ctx.Model(&insertData).Create(); err != nil {
				return fmt.Errorf("create3 fail: %v\n", err)
			}
			t.Logf("create3 success\n")
			return fmt.Errorf("不要提交这2个")
		})

		i++
		insertData.User = fmt.Sprintf("user%v", i)
		if err := ctx.Model(&insertData).Create(); err != nil {
			return fmt.Errorf("create4 fail: %v\n", err)
		}
		t.Logf("create success\n")

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

}
