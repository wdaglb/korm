# korm
golang orm, 一个简单易用的orm, 支持嵌套事务

## 安装
```
go get github.com/wdaglb/korm
go get github.com/go-sql-driver/mysql
```

## 支持数据库

* mysql https://github.com/go-sql-driver/mysql
* mssql https://github.com/denisenkom/go-mssqldb
* ...其它未测

## 连接mysql数据库

```
err := Connect(Config{
    Driver: "mysql",
    Host:   "127.0.0.1",
    Port:   3306,
    User:   "root",
    Pass:   "",
    Database: "test",
})
if err != nil {
    t.Fatal(err)
}
```

## 连接上下文
数据库的读写操作都依托于Context类
Context内部会自动维护db连接，不需要你自行管理Context实例，每次使用都建议实例一个新的Context
```
ctx := NewContext()
```

## 声明模型结构
```
type Test struct {
	Id int64 `db:"id"`
	User string `db:"user"`
}
```

## 查询一行数据
```
row := &Test{}
if ok, err := ctx.Model(&row).Where("Id", 1).Find(); !ok || err != nil {
    fmt.Println("记录不存在")
    return
}
fmt.Printf("id: %d\n", row.Id)
```
执行的sql
```
SELECT * FROM test WHERE id=1
```

## 查询多行数据
```
var rows []Test
if err := ctx.Model(&rows).Select(); err != nil {
    fmt.Println("查询错误")
}
fmt.Printf("rows: %v\n", rows)
```
执行的sql
```
SELECT * FROM test
```

## 创建数据
```
insertData := &Test{
  User: "test",
}
if err := ctx.Model(&insertData).Create(); err != nil {
    fmt.Println("创建错误")
}
fmt.Printf("insertData: %v\n", insertData)
```
执行的sql
```
INSERT INTO test (`user`) VALUES ('test')
```

## 更新数据
```
updateData := &Test{
  Id: 1,
  User: "test",
}
if err := ctx.Model(&updateData).Update(); err != nil {
    fmt.Println("更新错误")
}
fmt.Printf("updateData: %v\n", updateData)
```
执行的sql
```
UPDATE test SET `user`='test' WHERE `id`=1
```

## 删除数据
```
if err := ctx.Model(&Test{Id: 1}).Delete(); err != nil {
    fmt.Println("删除错误")
}
fmt.Printf("删除成功\n")
```
执行的sql
```
DELETE FROM test WHERE `id`=1
```

## 事务操作
```
ctx.Transaction(func () error {
    // 事务逻辑代码
})
```
只需要把需要进行的事务，写到闭包函数里即可，支持嵌套事务
注意：在同一个Context实例里的才会被事务影响

## 一对一关联

在模型定义声明字段
使用*指针标识，则需要With("模型名")加载关联数据

未使用*指针标识，则默认自动加载关联数据

```
type Test struct {
	Id int64 `db:"id"`
	Cate TestCate `pk:"Id" fk:"TestId"`
	Cate2 *TestCate `pk:"Id" fk:"TestId"`
}
type TestCate struct {
	Id int64 `db:"id"`
	Name string `db:"name"`
	TestId int `db:"test_id"`
}
```

## 一对多关联

在模型定义声明一个分片字段

```
type Test struct {
	Id int64 `db:"id"`
	Cates []TestCate `pk:"Id" fk:"TestId"`
}
type TestCate struct {
	Id int64 `db:"id"`
	Name string `db:"name"`
	TestId int `db:"test_id"`
}
```

## License
@ King east, 2021-now

Released under the [MIT License](https://github.com/wdaglb/korm/blob/main/LICENSE)
