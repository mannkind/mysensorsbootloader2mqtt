package main

const (
	idRequestTopic                 string = "%s/255/255/3/0/3"
	idResponseTopic                string = "%s/255/255/3/0/4"
	firmwareConfigRequestTopic     string = "%s/+/255/4/0/0"
	firmwareConfigResponseTopic    string = "%s/%s/255/4/0/1"
	firmwareRequestTopic           string = "%s/+/255/4/0/2"
	firmwareResponseTopic          string = "%s/%s/255/4/0/3"
	firmwareBootloaderCommandTopic string = "mysensors/bootloader/+/+"
	firmwareBlockSize              uint16 = 16
)
