package model

type ClientInfo struct {
	ID       string `gorm:"varchar(32)"`
	WanIP    string `gorm:"varchar(46)"`
	Hostname string `gorm:"varchar(64)"`
	OS       string `gorm:"varchar(32)"`
	Pickup   int8   // -1被标记为非活动 1表示申请捡起,2表示已经捡起
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
