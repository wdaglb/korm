package korm

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/wdaglb/korm/sqltype"
	"os"
	"strconv"
	"testing"
)

type Test struct {
	Id int64 `db:"id"`
	User string `db:"user"`
	TestId int `db:"test_id"`
	CreateTime sqltype.Timestamp `db:"create_time"`
	UpdateTime sqltype.DateTime `db:"update_time"`
	Cate *TestCate `pk:"Id" fk:"TestId"`
	Cates []TestCate `pk:"Id" fk:"TestId"`
}

type TestCate struct {
	Id int64 `db:"id"`
	Name string `db:"name"`
	TestId int `db:"test_id"`
}

func init()  {
	_ = godotenv.Load(".env")
	val, _ := strconv.Atoi(os.Getenv("DB_PORT"))

	conn := NewConnect(Config{MaxOpenConns: 100, MaxIdleConns: 10, PrintSql: true})
	err := conn.AddDb(DbConfig{
		Conn: "default",
		Driver: os.Getenv("DB_DRIVER"),
		Host:   os.Getenv("DB_HOST"),
		Port:   val,
		User:   os.Getenv("DB_USER"),
		Pass:   os.Getenv("DB_PASS"),
		Database: os.Getenv("DB_DATABASE"),
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

// 测试统计查询
func TestCount(t *testing.T)  {
	ctx := NewContext()
	count, _ := ctx.Model(Test{}).Count()

	t.Logf("count: %v\n", count)
}

// 测试求和查询
func TestSum(t *testing.T)  {
	ctx := NewContext()
	dst := 0
	_ = ctx.Model(Test{}).Sum("Id", &dst)

	t.Logf("sum: %v\n", dst)
}

// 测试最大值查询
func TestMax(t *testing.T)  {
	ctx := NewContext()
	dst := 0
	_ = ctx.Model(Test{}).Max("Id", &dst)

	t.Logf("max: %v\n", dst)
}

// 测试最小值查询
func TestMin(t *testing.T)  {
	ctx := NewContext()
	dst := 0
	_ = ctx.Model(Test{}).Min("Id", &dst)

	t.Logf("min: %v\n", dst)
}

// 测试平均值查询
func TestAvg(t *testing.T)  {
	ctx := NewContext()
	var dst float64
	_ = ctx.Model(Test{}).Avg("Id", &dst)

	t.Logf("avg: %v\n", dst)
}

// 测试数据创建
func TestCreate(t *testing.T)  {
	ctx := NewContext()

	cates := make([]TestCate, 0)
	cates = append(cates, TestCate{
		Name: "分类1",
	})
	cates = append(cates, TestCate{
		Name: "分类2",
	})

	insertData := &Test{
		User: "test",
		Cates: cates,
	}

	if err := ctx.Model(&insertData).Create(); err != nil {
		t.Fatalf("create fail: %v\n", err)
	}
	if insertData.Id <= 0 {
		t.Fatalf("create fail, id: %v\n", insertData.Id)
	}
	assert.Equal(t, insertData.User, "test")
}

// 测试数据更新
func TestUpdate(t *testing.T)  {
	ctx := NewContext()
	row := &Test{}

	if err := ctx.Model(&row).Find().Error; err != nil {
		t.Fatalf("getdata fail: %v\n", err)
	}
	row.User = "testUpdate"
	// row.Cate.Name = "2xx3"
	if err := ctx.Model(&row).Update(); err != nil {
		t.Fatalf("update fail: %v\n", err)
	}
	assert.Equal(t, row.User, "testUpdate")
}

// 测试多行查询
func TestSelect(t *testing.T)  {
	ctx := NewContext()
	var rows []Test
	// Where("Id", "in", []int{1, 2, 3, 4}).
	if err := ctx.Model(&rows).With("Cates").OrderByDesc("Id").Limit(3).Select().Error; err != nil {
		t.Fatalf("select fail: %v", err)
	}
	fmt.Printf("rows id: %d\n", rows[0].Id)
	if rows[0].Cate != nil {
		fmt.Printf("rows cate: %v\n", rows[0].Cate)
	}
	// fmt.Printf("rows cate2: %v\n", rows[0].Cate2.Name)

	fmt.Printf("rows cates: %v\n", rows[0].Cates)
}

// 测试单行查询
func TestFind(t *testing.T)  {
	ctx := NewContext()

	row := &Test{}

	if !ctx.Model(&row).With("Cates", func(db *Model) {
		db.OrderByDesc("Id").Limit(2)
	}).Find().Exist {
		t.Fatalf("记录不存在")
	}
	fmt.Printf("row cate: %d, %v\n", row.Id, row.Cate)

	fmt.Printf("row cates: %d, %v\n", row.Id, row.Cates)
}

// 测试数据删除
func TestDelete(t *testing.T)  {
	ctx := NewContext()
	row := &Test{}

	if !ctx.Model(&row).Find().Exist {
		t.Fatalf("记录不存在\n")
	}
	fmt.Printf("cates; %v, %v\n", row.Id, row.Cates)
	if err := ctx.Model(&row).Delete(); err != nil {
		t.Fatalf("delete fail: %v\n", err)
	}
}
