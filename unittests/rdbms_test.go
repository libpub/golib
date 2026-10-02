package unittests

import (
	"os"
	"testing"

	"github.com/libpub/golib/definations"
	"github.com/libpub/golib/orm/rdbms"
	"github.com/libpub/golib/orm/rdbms/behaviors"
	"github.com/libpub/golib/testingutil"
	"github.com/libpub/golib/utils"
)

const (
	testingDB = "default"
)

// User 内置用户模型
type User struct {
	ID                          int64  `xorm:"'id' BigInt pk autoincr" json:"id" form:"id"`
	Name                        string `xorm:"'name' VARCHAR(64) notnull index" json:"name" form:"name"`
	Telephone                   string `xorm:"'telephone' VARCHAR(64) null index" json:"telephone" form:"telephone"`
	Email                       string `xorm:"'email' VARCHAR(64) null index" json:"email" form:"email"`
	Avatar                      string `xorm:"'avatar' VARCHAR(64) null index" json:"avatar" form:"avatar"`
	HashedPassword              string `xorm:"'passwd_hash' VARCHAR(64)" json:"-" form:"-"`
	behaviors.ModifyingBehavior `xorm:"extends"`
	rdbms.Datasource            `xorm:"-" datasource:"default"`
}

// TableName User table name
func (u *User) TableName() string {
	return "sys_user"
}

func TestRDBMS_Orm(t *testing.T) {
	initRDBMSTestingDB(t)
	user := User{}
	err := rdbms.GetInstance().EnsureTableStructures(&user)
	testingutil.AssertNil(t, err, "EnsureTableStructures for user failed")
	rows, err := rdbms.GetInstance().FetchAll(&user)
	testingutil.AssertNil(t, err, "rdbms.FetchAll")
	testingutil.AssertNotNil(t, rows, "rdbms.FetchAll rows")
	_, err = rdbms.GetInstance().FetchAll(&user)
	testingutil.AssertNil(t, err, "rdbms.FetchAll")
}

func TestRDBMS_DynamicModel(t *testing.T) {
	initRDBMSTestingDB(t)
	jsonDefinition := `{"tableName": "users", "columns": [{"name": "id", "type": "int", "primaryKey": true, "autoIncrement": true}, {"name": "name", "type": "varchar(255)", "index": true}, {"name": "age", "type": "int"}, {"name": "rate", "type": "float"}], "indexes": [{"name": "idx_name", "columns": ["name"]}]}`
	tableDefination, err := rdbms.ParseDynamicModelFromJSON(jsonDefinition)
	testingutil.AssertNil(t, err, "ParseDynamicModelFromJSON for users table failed")
	userModel := rdbms.CreateDynamicModel(tableDefination)
	err = rdbms.GetInstance().EnsureTableStructures(&userModel)
	testingutil.AssertNil(t, err, "EnsureTableStructures for user failed")
	rows, err := rdbms.GetInstance().FetchAll(&userModel)
	testingutil.AssertNil(t, err, "rdbms.FetchAll")
	testingutil.AssertNotNil(t, rows, "rdbms.FetchAll rows")
	_, err = rdbms.GetInstance().FetchAll(&userModel)
	testingutil.AssertNil(t, err, "rdbms.FetchAll")
}

func initRDBMSTestingDB(t *testing.T) {
	dbFile := "./testing.db.test"
	dbConfig := definations.DBConnectorConfig{
		Driver:  "sqlite3",
		Address: "file://" + dbFile,
		Db:      testingDB,
	}
	if utils.IsPathExists(dbFile) {
		err := os.Remove(dbFile)
		testingutil.AssertNil(t, err, "Clean RDBMS testing db error")
	}
	engine, err := rdbms.GetInstance().Init(testingDB, &dbConfig)
	testingutil.AssertNil(t, err, "Init RDBMS error")
	testingutil.AssertNotNil(t, engine, "Init RDBMS engine")
}
