package model

type ClientInfo struct {
	ID       string `gorm:"varchar(37)"`
	WanIP    string `gorm:"varchar(46)"`
	Kernel   string `gorm:"varchar(25)"`
	OsInfo   string `gorm:"varchar(64)"`
	Pickup   int8   // -1被标记为非活动 1表示申请捡起,2表示已经捡起
	Timeout  string `gorm:"varchar(20)"` //捡起来多久才放下，例如2006-01-02 15:04:05
	Interval int32
	StartAt  int32
	//UpdateAt int64
}

//func (ci *ClientInfo) BeforeCreate(scope *gorm.Scope) (err error) {
//	//scope.SetColumn("CreateAt", time.Now().Unix())
//	scope.SetColumn("UpdateAt", time.Now().Unix())
//	return nil
//}
//
//func (ci *ClientInfo) BeforeUpdate(scope *gorm.Scope) (err error) {
//	scope.SetColumn("UpdateAt", time.Now().Unix())
//	return nil
//}
