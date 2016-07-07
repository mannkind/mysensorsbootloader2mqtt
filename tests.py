from firmware import * 
import unittest
from unittest.mock import Mock

class TestMySensorsFirmware(unittest.TestCase):
    def test_properties(self):
        firmware = MySensorsFirmware('files/test.hex')
        self.assertEqual(firmware.blocks, 80)
        self.assertEqual(firmware.crc, 18132)
        self.assertTrue(len(firmware.fwdata) > 0)

class TestMySensorsOTA(unittest.TestCase):
    def setUp(self):
        self.config = { 
            'mqtt': {
                'host': 'themis',
                'port': 1883,
                'sub_topic': 'test_sub_topic',
                'pub_topic': 'test_pub_topic'
            },
            'ota': {
                'update_blocks': 20,
                'types': {
                    1: 'Test'
                },
                'versions': {
                    1: '1.0'
                },
                'firmware': {
                    1: { 1: 'files/test.hex' }
                }
            }
        }
        self.mqtt = Mock()
    def test_init_no_auto_id(self):
        ota = MySensorsOTA(self.config, self.mqtt)

        self.mqtt.connect.assert_called_once_with(self.config['mqtt']['host'], self.config['mqtt']['port'], 60)
        self.assertEqual(self.mqtt.subscribe.call_count, 2)
        self.assertEqual(self.mqtt.message_callback_add.call_count, 3)
        self.assertEqual(ota.sub_topic, self.config['mqtt']['sub_topic'])
        self.assertEqual(ota.pub_topic, self.config['mqtt']['pub_topic'])

    def test_init_with_auto_id(self):
        self.config['auto_id'] = { 'next_id': 120 }
        ota = MySensorsOTA(self.config, self.mqtt)

        self.mqtt.connect.assert_called_once_with(self.config['mqtt']['host'], self.config['mqtt']['port'], 60)
        self.assertEqual(self.mqtt.subscribe.call_count, 3)
        self.assertEqual(self.mqtt.message_callback_add.call_count, 4)
        self.assertEqual(ota.sub_topic, self.config['mqtt']['sub_topic'])
        self.assertEqual(ota.pub_topic, self.config['mqtt']['pub_topic'])
        self.assertEqual(ota.next_id, self.config['auto_id']['next_id'])

    def test_on_id_request(self):
        self.config['auto_id'] = { 'next_id': 120 }
        message = Mock(topic = 'test_sub_topic/255/255/3/0/3', payload = b'')
        ota = MySensorsOTA(self.config, self.mqtt)
        (topic, payload) = ota.on_id_request(None, None, message)

        self.assertTrue(topic, 'test_sub_topic/255/255/3/0/4')
        self.assertTrue(payload, str(self.config['auto_id']['next_id'] + 1))

    def test_on_firmware_config_request(self):
        message = Mock(topic = 'test_pub_topic/1/255/4/0/0', payload = b'010001005000D446')
        ota = MySensorsOTA(self.config, self.mqtt)
        (topic, payload) = ota.on_firmware_config_request(None, None, message)

        self.assertEqual(topic, 'test_pub_topic/1/255/4/0/1')
        self.assertEqual(payload, '010001005000D446')

    def test_on_firmware_request(self):
        payloads = [
            '010001004F00FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF',
            '010001004E00FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF',
            '010001004D00FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF',
            '010001004C00FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF',
            '010001004B00FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF',
            '010001004A00FFCFFFFFFFFFFFFFFFFFFFFFFFFFFFFF',
            '0100010049008B002097E1F30E940000F9CF0895F894',
            '0100010048006B010E943E020E947000C0E0D0E00E94',
            '0100010047000F90DF91CF911F910F91089508950E94',
            '010001004600611103C01095812301C0812B8C939FBF',
            '010001004500FF1FEC55FF4FA591B4919FB7F8948C91',
            '01000100440021F069830E94A6016981E02FF0E0EE0F',
            '0100010043001491F901E057FF4F04910023C9F08823',
            '01000100420030E0F901E859FF4F8491F901E458FF4F',
            '0100010041000F931F93CF93DF931F92CDB7DEB7282F',
            '010001004000F8948C91822B8C939FBFDF91CF910895',
            '010001003F00322F309583238C938881822B888304C0',
            '010001003E008C93888182230AC0623051F4F8948C91',
            '010001003D00D4919FB7611108C0F8948C9120958223',
            '010001003C00E255FF4FA591B4918C559F4FFC01C591',
            '010001003B00FF4F8491882349F190E0880F991FFC01',
            '010001003A00DF9390E0FC01E458FF4F2491FC01E057',
            '01000100390003C08091B0008F7D8093B0000895CF93',
            '01000100380002C084B58F7D84BD08958091B0008F77',
            '010001003700809180008F7780938000089584B58F77',
            '0100010036008830B9F08430D1F4809180008F7D03C0',
            '01000100350028F4813099F08230A1F008958730A9F0',
            '0100010034008081806880831092C1000895833081F0',
            '0100010033008460808380818260808380818E7F8083',
            '010001003200E0EBF0E0808181608083EAE7F0E08081',
            '010001003100808181608083E1EBF0E0808184608083',
            '010001003000808182608083808181608083E0E8F0E0',
            '010001002F00EEE6F0E0808181608083E1E8F0E01082',
            '010001002E00816084BD85B5826085BD85B5816085BD',
            '010001002D009F908F900895789484B5826084BD84B5',
            '010001002C0029F7DDCFFF90EF90DF90CF90BF90AF90',
            '010001002B0083E0981EA11CB11CC114D104E104F104',
            '010001002A0070F321E0C21AD108E108F10888EE880E',
            '010001002900681979098A099B09683E734081059105',
            '010001002800D104E104F104F1F00E944E020E940E01',
            '010001002700FF926B017C010E940E014B015C01C114',
            '01000100260008958F929F92AF92BF92CF92DF92EF92',
            '010001002500911D43E0660F771F881F991F4A95D1F7',
            '0100010024003FBF6627782F892F9A2F620F711D811D',
            '01000100230026B5A89B05C02F3F19F00196A11DB11D',
            '0100010022008091090190910A01A0910B01B0910C01',
            '01000100210080910701909108012FBF08953FB7F894',
            '0100010020001F9018952FB7F8946091050170910601',
            '010001001F00AF919F918F913F912F910F900FBE0F90',
            '010001001E00090190930A01A0930B01B0930C01BF91',
            '010001001D00A0910B01B0910C010196A11DB11D8093',
            '010001001C00A0930701B09308018091090190910A01',
            '010001001B00A11DB11D209304018093050190930601',
            '010001001A0020F40296A11DB11D05C029E8230F0396',
            '0100010019000701B09108013091040126E0230F2D37',
            '0100010018009F93AF93BF938091050190910601A091',
            '0100010017001F920F920FB60F9211242F933F938F93',
            '0100010016007E438105910508F4A8951F910F910895',
            '010001001500020130910301601B710B820B930B6038',
            '01000100140031010E94020100910001109101012091',
            '0100010013008DE00E94080268EE73E080E090E00E94',
            '010001001200080268EE73E080E090E00E94310160E0',
            '0100010011006000A89508950F931F9361E08DE00E94',
            '01000100100090E00FB6F894A895809360000FBE2093',
            '010001000F0070930101809302019093030129E288E1',
            '010001000E0061E08DE00E94CF010E94020160930001',
            '010001000D00B207E1F70E943F020C944F020C940000',
            '010001000C00DEBFCDBF21E0A0E0B1E001C01D92AD30',
            '010001000B000000240027002A0011241FBECFEFD8E0',
            '010001000A000303030300000000250028002B000000',
            '01000100090004040404040404040202020202020303',
            '01000100080010204080010204081020010204081020',
            '01000100070000030407000000000000000001020408',
            '0100010006000C946E000C946E000000000800020100',
            '0100010005000C946E000C946E000C946E000C946E00',
            '0100010004000C94B8000C946E000C946E000C946E00',
            '0100010003000C946E000C946E000C946E000C946E00',
            '0100010002000C946E000C946E000C946E000C946E00',
            '0100010001000C946E000C946E000C946E000C946E00',
            '0100010000000C945C000C946E000C946E000C946E00'
        ]
        max = 80
        ota = MySensorsOTA(self.config, self.mqtt)
        for block in reversed(range(0, max)):
            block_as_hex = str(hex(block)).replace('0x','').zfill(2).upper()
            incoming = '01000100%s00' % block_as_hex
            message = Mock(topic = 'test_sub_topic/1/255/4/0/2', payload = incoming.encode())
            (topic, payload) = ota.on_firmware_request(None, None, message)

            self.assertEqual(topic, 'test_pub_topic/1/255/4/0/3')
            self.assertEqual(payload, payloads[max - block - 1])

    def test_on_bootloader_command(self):
        to = '1'
        command = '2'
        payload = '25'
        message = Mock(topic = 'mysbootloader_command/%s/%s' % (to, command), payload = payload.encode()) 
        ota = MySensorsOTA(self.config, self.mqtt)
        ota.on_bootloader_command(None, None, message)

        self.assertTrue(ota.bootloader_commands.get(to) is not None)
        self.assertEqual(ota.bootloader_commands.get(to)['command'], command)
        self.assertEqual(ota.bootloader_commands.get(to)['data'], payload)

        (topic, payload) = ota.run_bootloader_command(to)
        self.assertTrue(ota.bootloader_commands.get(to) is None)
        self.assertEqual(topic, 'test_pub_topic/1/255/4/0/1')
        self.assertEqual(payload, '0200190000007ADA')

if __name__ == '__main__':
    unittest.main()
