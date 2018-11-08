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
./mysb -c */the/path/to/config_folder/config.yaml*
```

# Configuration

Configuration happens in the config.yaml file. A full example might look this:

```
settings:
    clientid: 'GoMySysBootloader'
    broker:   'tcp://mosquitto:1883'
    subtopic: 'mysensors_rx'
    pubtopic: 'mysensors_tx'

control:
    autoidenabled: true   
    nextid: 12
    firmwarebasepath: '/the/path/to/config_folder/firmware'
    nodes:
        default: {
            type: 1,
            version: 1
        }
        1: { type: 1, version: 1 }
        2: { type: 3, version: 1 }
        3: { type: 1, version: 2 }
        4: { type: 1, version: 1 }
        5: { type: 2, version: 3 }
    # Not used in Mysb - for reference only
    types:
      1: 'Temperature Monitor'
      2: 'Door Monitor'
      3: 'Plant Monitor'
      4: 'Garage Actuator'
      5: 'Energy Monitor'
      6: 'Glass Break Monitor'
      7: 'Fireplace Actuator'
      8: 'Signing Setup'
      9: 'AC Actuator'

    versions:
      1: '1.5.1'
      2: '1.5.4'
      3: '2.0.0-beta'

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
