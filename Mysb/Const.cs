using System;

namespace Mysb
{
    public static class Const
    {
        public const string IdRequestTopicPartial = "255/255/3/0/3";
        public const string IdResponseTopicPartial = "255/255/3/0/4";
        public const string FirmwareConfigRequestTopicPartial = "+/255/4/0/0";
        public const string FirmwareConfigResponseTopicPartial = "255/4/0/1";
        public const string FirmwareRequestTopicPartial = "+/255/4/0/2";
        public const string FirmwareResponseTopicPartial = "255/4/0/3";
        public const string FirmwareBootloaderCommandTopicPartial = "mysensors/bootloader/+/+";
        public const UInt16 FirmwareBlockSize = 16;
    }
}