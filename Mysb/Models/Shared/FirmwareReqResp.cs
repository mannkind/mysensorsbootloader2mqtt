using System.Runtime.InteropServices;

namespace Mysb.Models.Shared
{
    /// <summary>
    /// 
    /// </summary>
    public struct FirmwareReqResp
    {
        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Type { get; set; }

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Version { get; set; }

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Block { get; set; }

        /// <summary>
        /// 
        /// </summary>
        [MarshalAs(UnmanagedType.ByValArray, SizeConst = Const.FirmwareBlockSize)]
        public byte[] Data;

        /// <inheritdoc />
        public override string ToString() =>
            $"Type: {this.Type.ToString()}, Version: {this.Version.ToString()}, " +
            $"Block: {(this.Block + 1).ToString()}";
    }
}
