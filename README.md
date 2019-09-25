# mysensorsbootloader2mqtt

[![Software
License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/mannkind/mysensorsbootloader2mqtt/blob/master/LICENSE.md)
[![Travis CI](https://img.shields.io/travis/mannkind/mysensorsbootloader2mqtt/master.svg?style=flat-square)](https://travis-ci.org/mannkind/mysensorsbootloader2mqtt)
[![Coverage Status](https://img.shields.io/codecov/c/github/mannkind/mysensorsbootloader2mqtt/master.svg)](http://codecov.io/github/mannkind/mysensorsbootloader2mqtt?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mannkind/mysensorsbootloader2mqtt)](https://goreportcard.com/report/github.com/mannkind/mysensorsbootloader2mqtt)

A Firmware Uploading Tool for the MySensors Bootloader via MQTT

## Installation

### Via Docker

```bash
docker run -d --name="mysensorsbootloader2mqtt" -v /etc/localtime:/etc/localtime:ro mannkind/mysensorsbootloader2mqtt
```

### Via Mage

```bash
git clone https://github.com/mannkind/mysensorsbootloader2mqtt
cd mysensorsbootloader2mqtt
mage
./mysensorsbootloader2mqtt
```

## Configuration

Configuration happens via environmental variables

```bash
MYSENSORS_SUBTOPIC           - [OPTIONAL] The MQTT topic on which to subscribe, defaults to "mysensors_rx"
MYSENSORS_PUBTOPIC           - [OPTIONAL] The MQTT topic on which to publish, defaults to "mysensors_tx"
MYSENSORS_AUTOID             - [OPTIONAL] The flag that indicates MySensorsBootloader should handle ID requests, defaults to false
MYSENSORS_NEXTID             - [OPTIONAL] The number on which to base the next id, defaults to 1
MYSENSORS_FIRMWAREBASEPATH   - [OPTIONAL] The path to the firmware files, defaults to "/config/firmware"
MYSENSORS_NODES              - [OPTIONAL] The nodes configuration (see below)
MQTT_CLIENTID                - [OPTIONAL] The clientId, defaults to "DefaultMySensorsBootloaderClientID"
MQTT_BROKER                  - [OPTIONAL] The MQTT broker, defaults to "tcp://mosquitto.org:1883"
MQTT_USERNAME                - [OPTIONAL] The MQTT username, default to ""
MQTT_PASSWORD                - [OPTIONAL] The MQTT password, default to ""
```

The file referenced by MYSENSORS_NODES might look something like the following

```yaml
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

1. A type/version _assigned_ to the node in the config.yaml file
2. The requested type/version sent in the configuration request
3. The default firmware setup in the config.yaml file

The location of the firmware picked is relative to the `control['firmwarebasepath']` setting in config.yaml.

E.g. /path/to/config*folder/firmware/\_type*/_version_/firmware.hex

```bash
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
