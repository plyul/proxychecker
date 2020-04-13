package checker

type ConfigStruct struct {
	ProxyListFile        string
	CheckURL             string
	NumWorkers           int
	MaxProxiesToCheck    int
	MaxAddressesToOutput int
	TgAPIToken           string
	TgProxy              string
}
