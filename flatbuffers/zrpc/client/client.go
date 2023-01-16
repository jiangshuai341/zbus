package client

//message
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

func BindDelegate(serviceType int32, funcHash int32, entityID int64, callback func([]byte, error)) {

}
func RemoteInvokeSync(serviceType int32, funcHash int32, entityID int64, arg []byte) {

}
func RemoteInvokeAsync(serviceType int32, funcHash int32, entityID int64, arg []byte, callback func([]byte, error)) {

}
