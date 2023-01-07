package toolkit

func init() {
	if initNetinterfaceIpMap() != nil {
		panic("initNetinterfaceIpMap Failed")
	}
}
