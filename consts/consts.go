package consts

import "gorm.io/gorm"

// DB 在服务启动时初始化，运行期间不替换；请求和事务使用各自的会话。
var DB *gorm.DB
