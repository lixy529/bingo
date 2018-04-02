// model demo
//   变更历史
//     2017-02-10  lixiaoya  新建
package models

import (
	"github.com/lixy529/bingo"
	_ "github.com/go-sql-driver/mysql"
)

type BaseModel struct {
	bingo.Model
}
