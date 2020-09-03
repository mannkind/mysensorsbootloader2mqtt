using System.Collections.Generic;
using System.Linq;

namespace Mysb.Models.Shared
{
    /// <summary>
    /// 
    /// </summary>
    public record Firmware
    {
        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Blocks { get; init; }

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Crc { get; init; }

        /// <summary>
        /// 
        /// </summary>
        /// <typeparam name="byte"></typeparam>
        /// <returns></returns>
        public IEnumerable<byte> Data { get; init; } = new List<byte>();

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public byte[] this[ushort key]
        {
            get => this.Data.
                Skip(key * Const.FirmwareBlockSize).
                Take(Const.FirmwareBlockSize).
                ToArray();
        }
    }
}
