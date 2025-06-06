package config

type monitoringNetworkType struct {
	Times int
	Runtime  int
}

var MonitoringNetwork = monitoringNetworkType{
	Times:1,  //checking time
	Runtime:2, //seconds
}
