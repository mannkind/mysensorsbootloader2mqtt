# Mysb

[![Software
License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/mannkind/mysb/blob/master/LICENSE.md)
[![Travis CI](https://img.shields.io/travis/mannkind/mysb/master.svg?style=flat-square)](https://travis-ci.org/mannkind/mysb)
[![Coverage Status](https://img.shields.io/codecov/c/github/mannkind/mysb/master.svg)](http://codecov.io/github/mannkind/mysb?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mannkind/mysb)](https://goreportcard.com/report/github.com/mannkind/mysb)

A Firmware Uploading Tool for the MYSBootloader via MQTT

# Installation

## Via Docker
```
docker run -d --name="mysb" -v /the/path/to/config_folder:/config -v /etc/localtime:/etc/localtime:ro mannkind/mysb
```

## Via Make
```
git clone https://github.com/mannkind/mysb
cd mysb
make
MYSB_CONFIGFILE="config.yaml" ./mysb 
```

# Configuration

Configuration happens via environmental variables

```
MYSB_AUTOID - The flag that indicates Mysb should handle ID requests, defaults to false
MYSB_NEXTID - The number on which to base the next id, defaults to 1
MYSB_FIRMWAREBASEPATH - The path to the firmware files, defaults to "/config/firmware"
MYSB_CONFIG - The yaml config that contains control variables for Mysb
MYSB_NODES - The nodes configuration (see below)
MQTT_CLIENTID - [OPTIONAL] The clientId, defaults to "DefaultMysbClientID"
MQTT_BROKER - [OPTIONAL] The MQTT broker, defaults to "tcp://mosquitto.org:1883"
MQTT_SUBTOPIC - [OPTIONAL] The MQTT topic on which to subscribe, defaults to "mysensors_rx"
MQTT_PUBTOPIC - [OPTIONAL] The MQTT topic on which to publish, defaults to "mysensors_tx"
MQTT_USERNAME - [OPTIONAL] The MQTT username, default to ""
MQTT_PASSWORD - [OPTIONAL] The MQTT password, default to ""
```

The file referenced by MYSB_NODES might look something like the following

```
default: {
    type: 1,
    version: 1
}
1: { type: 1, version: 1 }
2: { type: 3, version: 1 }
3: { type: 1, version: 2 }
4: { type: 1, version: 1 }
5: { type: 2, version: 3 }
```

The firmware a node is using is a combination of a `type` and a `version`. The priority of the firmware used is based on the following:

1. A type/version *assigned* to the node in the config.yaml file
2. The requested type/version sent in the configuration request
3. The default firmware setup in the config.yaml file

The location of the firmware picked is relative to the `control['firmwarebasepath']` setting in config.yaml.

E.g. /path/to/config_folder/firmware/_type_/_version_/firmware.hex

```
$ find /path/to/config_folder/firmware
/path/to/config_folder/firmware/3
/path/to/config_folder/firmware/3/1
/path/to/config_folder/firmware/3/1/firmware.hex
/path/to/config_folder/firmware/2
/path/to/config_folder/firmware/2/1
/path/to/config_folder/firmware/2/1/firmware.hex
/path/to/config_folder/firmware/2/2
/path/to/config_folder/firmware/2/2/firmware.hex
/path/to/config_folder/firmware/2/3
/path/to/config_folder/firmware/2/3/firmware.hex
/path/to/config_folder/firmware/1
/path/to/config_folder/firmware/1/1
/path/to/config_folder/firmware/1/1/firmware.hex
/path/to/config_folder/firmware/1/2
/path/to/config_folder/firmware/1/2/firmware.hex
```
