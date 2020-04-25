# mysensorsbootloader2mqtt

[![Software
License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/mannkind/mysensorsbootloader2mqtt/blob/master/LICENSE.md)
[![Build Status](https://github.com/mannkind/mysensorsbootloader2mqtt/workflows/Main%20Workflow/badge.svg)](https://github.com/mannkind/mysensorsbootloader2mqtt/actions)
[![Coverage Status](https://img.shields.io/codecov/c/github/mannkind/mysensorsbootloader2mqtt/master.svg)](http://codecov.io/github/mannkind/mysensorsbootloader2mqtt?branch=master)

A Firmware Uploading Tool for the MySensors Bootloader via MQTT

## Use

The application can be locally built using `dotnet build` or you can utilize the multi-architecture Docker image(s).

### Example

```bash
docker run \
-e MYSB__AUTOIDENABLED="true" \
-e MYSB__NEXTID="12" \
-e MYSB__RESOURCES__0__NodeID="default" \
-e MYSB__RESOURCES__0__Type="1" \
-e MYSB__RESOURCES__0__Version="1" \
-e MYSB__RESOURCES__1__NodeID="1" \
-e MYSB__RESOURCES__1__Type="4" \
-e MYSB__RESOURCES__1__Version="2" \
-e MYSB__MQTT__BROKER="localhost" \
mannkind/litterrobot2mqtt:latest
```

OR

```bash
MYSB__AUTOIDENABLED="true" \
MYSB__NEXTID="12" \
MYSB__RESOURCES__0__NodeID="default" \
MYSB__RESOURCES__0__Type="1" \
MYSB__RESOURCES__0__Version="1" \
MYSB__RESOURCES__1__NodeID="1" \
MYSB__RESOURCES__1__Type="4" \
MYSB__RESOURCES__1__Version="2" \
MYSB__MQTT__BROKER="localhost" \
./litterrobot2mqtt 
```


## Configuration

Configuration happens via environmental variables

```bash
MYSB__AUTOIDENABLED                      - [OPTIONAL] The flag that indicates MySensorsBootloader should handle ID requests, defaults to false
MYSB__NEXTID                             - [OPTIONAL] The number on which to base the next id, defaults to 1
MYSB__FIRMWAREBASEPATH                   - [OPTIONAL] The path to the firmware files, defaults to "/config/firmware"
MYSB__RESOURCES__#__NodeId               - [OPTIONAL] The nodes configuration NodeId
MYSB__RESOURCES__#__Type                 - [OPTIONAL] The nodes configuration Type
MYSB__RESOURCES__#__Version              - [OPTIONAL] The nodes configuration Version
MYSB__SUBTOPIC                           - [OPTIONAL] The MQTT topic on which to subscribe, defaults to "mysensors_rx"
MYSB__PUBTOPIC                           - [OPTIONAL] The MQTT topic on which to publish, defaults to "mysensors_tx"
MYSB__MQTT__BROKER                       - [OPTIONAL] The MQTT broker, defaults to "test.mosquitto.org"
MYSB__MQTT__USERNAME                     - [OPTIONAL] The MQTT username, default to ""
MYSB__MQTT__PASSWORD                     - [OPTIONAL] The MQTT password, default to ""
```

`MYSB__RESOURCES` is a list of objects that have a NodeId, Type, and Version.

```bash
MYSB__RESOURCES__0__NodeID="default"
MYSB__RESOURCES__0__Type="1"
MYSB__RESOURCES__0__Version="1"
MYSB__RESOURCES__1__NodeID="1"
MYSB__RESOURCES__1__Type="1"
MYSB__RESOURCES__1__Version="1"
MYSB__RESOURCES__2__NodeID="2"
MYSB__RESOURCES__2__Type="3"
MYSB__RESOURCES__2__Version="1"
MYSB__RESOURCES__3__NodeID="3"
MYSB__RESOURCES__3__Type="1"
MYSB__RESOURCES__3__Version="2"
```

The firmware a node is using is a combination of a `type` and a `version`. The priority of the firmware used is based on the following:

* Load the user-defined firmware 
* Load the node-defined firmware
* Load the user-defined default firmware


E.g. /path/to/config\_folder/firmware/`type`/`version`/firmware.hex

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
