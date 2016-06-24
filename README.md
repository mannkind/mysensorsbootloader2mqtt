# Mysb

[![Software
License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/mannkind/mysb/blob/master/LICENSE.md)
[![Travis
CI](https://img.shields.io/travis/mannkind/mysb/master.svg?style=flat-square)](https://travis-ci.org/mannkind/mysb)
[![Coverage Status](http://codecov.io/github/mannkind/mysb/coverage.svg?branch=master)](http://codecov.io/github/mannkind/mysb?branch=master)

A Firmware Uploading Tool for the MYSBootloader via MQTT

# Configuration

Configuration happens in the config.yaml file. A full example might look this:

```
settings:
    clientid: 'GoMySysBootloader'
    broker:   'tcp://mosquitto:1883'
    subtopic: 'mysensors_rx'
    pubtopic: 'mysensors_tx'

control:
    nextid: 12
    firmwarebasepath: '/Data/mysb/firmware'
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

E.g. *control['firmwarebasepath']*/*type*/*version*/firmware.hex

```
$ find /Data/mysb/firmware
/Data/mysb/firmware/3
/Data/mysb/firmware/3/1
/Data/mysb/firmware/3/1/firmware.hex
/Data/mysb/firmware/2
/Data/mysb/firmware/2/1
/Data/mysb/firmware/2/1/firmware.hex
/Data/mysb/firmware/2/2
/Data/mysb/firmware/2/2/firmware.hex
/Data/mysb/firmware/2/3
/Data/mysb/firmware/2/3/firmware.hex
/Data/mysb/firmware/1
/Data/mysb/firmware/1/1
/Data/mysb/firmware/1/1/firmware.hex
/Data/mysb/firmware/1/2
/Data/mysb/firmware/1/2/firmware.hex
```
