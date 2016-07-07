import codecs
import struct
import yaml
import paho.mqtt.client as paho_mqtt

FIRMWARE_BLOCK_SIZE = 16


class MySensorsOTA:
    def __init__(self, config, mqtt):
        self.mqtt_config = config.get('mqtt', {})
        self.ota_config = config.get('ota', {})
        self.autoid_config = config.get('auto_id', {})

        self.update_blocks = self.ota_config.get('update_blocks', 100)
        self.next_id = self.autoid_config.get('next_id')
        self.bootloader_commands = {}

        self.sub_topic = self.mqtt_config.get('sub_topic', 'mysensors_rx')
        self.pub_topic = self.mqtt_config.get('pub_topic', 'mysensors_tx')
        self.host = self.mqtt_config.get('host', 'localhost')
        self.port = self.mqtt_config.get('port', 1883)

        print('MySensorsOTA:')
        print('    MQTT: %s:%s' % (self.host, self.port))
        print('    Subscription Topic: %s' % self.sub_topic)
        print('    Publish Topic: %s' % self.pub_topic)
        print('    Next ID: %s' % self.next_id)
        print('')

        self.mqtt = mqtt
        self.mqtt.connect(self.host, self.port, 60)

        print("Connected to MQTT; subscribing to necessary topics.")
        if self.next_id is not None:
            self.mqtt.subscribe('%s/255/255/3/0/3' % self.sub_topic)
            self.mqtt.message_callback_add(
                '%s/255/255/3/0/3' % self.sub_topic,
                self.on_id_request
            )
        self.mqtt.subscribe('mysbootloader_command/+/+')
        self.mqtt.message_callback_add(
            'mysbootloader_command/+/+',
            self.on_bootloader_command
        )
        self.mqtt.subscribe('%s/+/255/4/0/+' % self.sub_topic)
        self.mqtt.message_callback_add(
            '%s/+/255/4/0/0' % self.sub_topic,
            self.on_firmware_config_request
        )
        self.mqtt.message_callback_add(
            '%s/+/255/4/0/2' % self.sub_topic,
            self.on_firmware_request
        )

    def loop(self):
        self.mqtt.loop_forever()

    def on_bootloader_command(self, client, userdata, message):
        topic = message.topic
        payload = message.payload.decode('ascii')

        (_, to, command) = topic.split('/')
        self.bootloader_commands[to] = {
            'command': command,
            'data': payload
        }

    def run_bootloader_command(self, to):
        blcmd = self.bootloader_commands.get(to)
        if blcmd is not None:
            self.bootloader_commands.pop(to)
            type = int(blcmd['command'])
            version = 0

            # 0x01 - Erase EEPROM
            # 0x02 - Set NodeID
            # 0x03 - Set ParentID
            if type == 0x02 or type == 0x03:
                version = int(blcmd['data'])

            resp_topic = '%s/%s/255/4/0/1' % (self.pub_topic, to)
            resp_pk = struct.pack('<HHHH', type, version, 0, 0xDA7A)
            resp_data = str(codecs.encode(resp_pk, 'hex'), 'ascii').upper()
            self.mqtt.publish(resp_topic, resp_data)

            return (resp_topic, resp_data)

        return False

    def on_id_request(self, client, userdata, message):
        self.next_id += 1
        resp_topic = '%s/255/255/3/0/4' % self.pub_topic
        self.mqtt.publish(resp_topic, self.next_id)

        return (resp_topic, self.next_id)

    def on_firmware_config_request(self, client, userdata, message):
        topic = message.topic
        payload = message.payload.decode('ascii')

        to = topic.split('/')[1]
        (type, version, blocks, crc) = \
            struct.unpack('<HHHH', bytearray.fromhex(payload[0:16]))

        # Attempt to run any bootloader commands
        if self.run_bootloader_command(to):
            return

        config = MySensorsOTAConfig(self.ota_config, to, type, version)
        firmware = MySensorsFirmware(config.filename)

        resp_topic = '%s/%s/255/4/0/1' % (self.pub_topic, to)
        resp_pk = struct.pack(
            "<HHHH", type, version, firmware.blocks, firmware.crc
        )
        resp_payload = str(codecs.encode(resp_pk, 'hex'), 'ascii').upper()

        self.mqtt.publish(resp_topic, resp_payload)

        return (resp_topic, resp_payload)

    def on_firmware_request(self, client, userdata, message):
        topic = message.topic
        payload = message.payload.decode('ascii')
        to = topic.split('/')[1]

        payload = payload[0:12]
        (type, version, blocks) = \
            struct.unpack('<HHH', bytearray.fromhex(payload))
        config = MySensorsOTAConfig(self.ota_config, to, type, version)
        firmware = MySensorsFirmware(config.filename)

        if blocks % self.update_blocks == 0:
            print(
                "Sending block %d of %s %s" %
                (blocks, config.type_name, config.version_name)
            )

        from_block = blocks * FIRMWARE_BLOCK_SIZE
        to_block = from_block + FIRMWARE_BLOCK_SIZE
        update_data = firmware.fwdata[from_block:to_block]

        resp_topic = '%s/%s/255/4/0/3' % (self.pub_topic, to)
        resp_hex = codecs.encode(
            struct.pack("<HHH", type, version, blocks),
            'hex'
        )
        resp_pk = str(resp_hex, 'ascii')
        data_hex = codecs.encode(update_data, 'hex')
        data = str(data_hex, 'ascii')

        resp_payload = ('%s%s' % (resp_pk, data)).upper()
        self.mqtt.publish(resp_topic, resp_payload)

        return (resp_topic, resp_payload)


class MySensorsOTAConfig:
    def __init__(self, config, node_id, r_type, r_version):
        self.type_name = 'Unknown'
        self.version_name = 'Unknown'
        try:
            a_type = config.get('nodes', {}).get(node_id, {}).get('type')
            a_version = config.get('nodes', {}).get(node_id, {}).get('version')
            self.type_name = config.get('types', {}).get(a_type, 'Unknown')
            self.version_name = config.get('versions', {}) \
                                      .get(a_version, 'Unknown')
            self.filename = config.get('firmware', {}) \
                                  .get(a_type, {}) \
                                  .get(a_version)
        except Exception as e:
            print(
                "MySensorsFirmwareConfig: !Assigned %s %s. %s" %
                (self.type_name, self.version_name, e)
            )

        if self.filename is not None:
            return

        try:
            self.type_name = config.get('types', {}).get(r_type, 'Unknown')
            self.version_name = config.get('versions', {}) \
                                      .get(r_version, 'Unknown')
            self.filename = config.get('firmware', {}) \
                                  .get(r_type, {}) \
                                  .get(r_version)
        except Exception as e:
            print(
                "MySensorsFirmwareConfig: !Requested %s %s. %s" %
                (self.type_name, self.version_name, e)
            )

        if self.filename is not None:
            return

        try:
            d_type = config.get('nodes', {}).get('default', {}).get('type')
            d_version = config.get('nodes', {}) \
                              .get('default', {}) \
                              .get('version')
            self.type_name = config.get('types', {}).get(d_type, 'Unknown')
            self.version_name = config.get('versions', {}) \
                                      .get(d_version, 'Unknown')
            self.filename = config.get('firmware') \
                                  .get(d_type, {}) \
                                  .get(d_version)
            return None
        except Exception as e:
            print(
                "MySensorsFirmwareConfig: !Default %s %s. %s" %
                (self.type_name, self.version_name, e)
            )


class MySensorsFirmware:
    def __init__(self, filename):
        self.blocks = 0
        self.fwdata = bytearray()
        self.crc = 0

        self.load(filename)

    def load(self, filename):
        start = 0
        end = 0
        with open(filename) as f:
            for line in f:
                line = line.strip()
                if len(line) == 0:
                    continue
                while line[0] != ':':
                    line = line[1:]

                rlen = int(line[1:3], 16)
                offset = int(line[3:7], 16)
                rtype = int(line[7:9], 16)
                data = line[9:9 + (2 * rlen)]

                if rtype != 0:
                    continue

                if start == 0 and end == 0:
                    start = offset
                    end = offset

                while offset > end:
                    self.fwdata.append(255)
                    end += 1

                for i in [[(x * 2), (x * 2) + 2] for x in range(0, rlen)]:
                    self.fwdata.append(int(data[i[0]:i[1]], 16))

                end += rlen

        pad = end % 128
        for i in range(0, 128 - pad):
            self.fwdata.append(255)
            end += 1

        self.blocks = int((end - start) / FIRMWARE_BLOCK_SIZE)
        self.crc = 0xFFFF
        for i in range(0, self.blocks * FIRMWARE_BLOCK_SIZE):
            crc = self.crc ^ (self.fwdata[i] & 0xFF)
            for j in range(0, 8):
                if (crc & 1) > 0:
                    crc = ((crc >> 1) ^ 0xA001)
                else:
                    crc = (crc >> 1)

            self.crc = crc

if __name__ == '__main__':
    config = yaml.load(open('firmware.yaml'))
    mqtt = paho_mqtt.Client()
    myota = MySensorsOTA(config, mqtt)
    myota.loop()
