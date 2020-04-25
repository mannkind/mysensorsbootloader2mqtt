using System.Collections.Generic;
using System.Linq;

namespace Mysb.Models.Shared
{
    /// <summary>
    /// 
    /// </summary>
    public class Firmware
    {
        public ushort Blocks { get; set; }
        public ushort Crc { get; set; }
        public IEnumerable<byte> Data { get; set; } = new List<byte>();

        public byte[] this[ushort key]
        {
            get => this.Data.
                Skip(key * Const.FirmwareBlockSize).
                Take(Const.FirmwareBlockSize).
                ToArray();
        }
    }
}
