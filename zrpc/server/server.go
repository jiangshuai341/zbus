package server

/*
0       2       4       6       8 (BYTE)
+-------+-------+-------+-------+
|     pakLen    |       cmd     |
+---------------+---------------+   8
|  serviceType  |    funcHash   |
+---------------+---------------+  16
|            entityID           |
+---------------+---------------+  24
|                               |
|              DATA             |
|                               |
+-------------------------------+
*/

func init() {
}

func ConnectToProxy() {
	//zrpccommon.RegisterServiceToETCD()
}

func CreateEntity(serviceType int32, key int32, entityID int64, imp any) {

}

func DeclareDelegate(serviceType int32, funcHashVal int32, entityID int64) {

}

func RegistService(serviceType int32, funcHashVal int32, entityID int64) {

}

func ExecuteSync(serviceType int32, funcHash int32, entityID int64, arg []byte) {

}

func ExecuteAsync(serviceType int32, funcHash int32, entityID int64, arg []byte, callback func([]byte, error)) {

}

func Broadcast(serviceType int32, funcHash int32, entityID int64, arg []byte, callback func([]byte, error)) {

}
